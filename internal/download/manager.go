package download

import (
	"fmt"
	"os"
	"sync"
	"torrentd/internal/peer"
	"torrentd/internal/torrentfile"
	"torrentd/internal/tracker"
)

type Manager struct {
	mu       sync.Mutex
	torrents map[string]*Torrent
	peerId   [20]byte
}

func NewManager() (*Manager, error) {
	peerId, err := tracker.GeneratePeerID()
	if err != nil {
		return nil, err
	}
	return &Manager{
		torrents: make(map[string]*Torrent),
		peerId:   peerId,
	}, nil
}

func (m *Manager) AddTorrent(path string) (string, error) {
	tf, err := torrentfile.Open(path)
	if err != nil {
		return "", err
	}

	t := &Torrent{
		Name:  tf.Name,
		Total: len(tf.PieceHashes),
	}
	t.setStatus("starting")

	id := fmt.Sprintf("%x", tf.InfoHash)
	m.mu.Lock()
	m.torrents[id] = t
	m.mu.Unlock()

	go func() {
		resp, err := tracker.Announce(tf, m.peerId, 6881)
		if err != nil {
			t.setStatus("error")
			return
		}
		err = Download(t, tf, resp.Peers, m.peerId)
		if err != nil {
			t.setStatus("error")
			return
		}

	}()
	return id, nil
}

func (m *Manager) List() []*Torrent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*Torrent, 0, len(m.torrents))
	for _, t := range m.torrents {
		out = append(out, t)

	}
	return out
}

func (m *Manager) Get(id string) (*Torrent, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.torrents[id]
	return t, ok
}

func Download(t *Torrent, tf *torrentfile.TorrentFile, peers []peer.Peer, peerID [20]byte) error {

	f, err := os.OpenFile(tf.Name, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Truncate(tf.Length); err != nil {
		return err
	}

	workQueue := make(chan pieceWork, len(tf.PieceHashes))
	results := make(chan pieceResult)
	t.setStatus("downloading")

	for index, hash := range tf.PieceHashes {
		start, end := tf.PieceBounds(index)
		length := end - start
		workQueue <- pieceWork{index: index, hash: hash, length: length}
	}

	for _, p := range peers {
		go worker(p, tf.InfoHash, peerID, workQueue, results)
	}

	var donePieces int
	for donePieces < len(tf.PieceHashes) {
		res := <-results
		start, _ := tf.PieceBounds(res.index)
		if _, err := f.WriteAt(res.buf, int64(start)); err != nil {
			t.setStatus("error")
			return err
		}
		donePieces++
		t.pieceDone()
	}
	t.setStatus("done")
	return nil
}
