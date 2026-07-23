package download

import (
	"context"
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
	wg       sync.WaitGroup
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

	ctx, cancel := context.WithCancel(context.Background())

	t := &Torrent{
		Name:   tf.Name,
		Total:  len(tf.PieceHashes),
		cancel: cancel,
	}
	t.setStatus("starting")

	id := fmt.Sprintf("%x", tf.InfoHash)
	m.mu.Lock()
	m.torrents[id] = t
	m.mu.Unlock()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		defer t.cancel()
		resp, err := tracker.Announce(tf, m.peerId, 6881)
		if err != nil {
			t.setStatus("error")
			return
		}
		err = Download(ctx, t, tf, resp.Peers, m.peerId)
		if err != nil {
			t.setStatus("error")
			return
		}

	}()
	return id, nil
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	for _, t := range m.torrents {
		t.cancel()
	}
	m.mu.Unlock()

	m.wg.Wait()
}

func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	t, ok := m.torrents[id]
	if ok {
		delete(m.torrents, id)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("torrent %s not found", id)

	}
	t.cancel()
	t.setStatus("removed")
	return nil
}

func (m *Manager) Pause(id string) error {
	m.mu.Lock()
	t, ok := m.torrents[id]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("torrent %s not found", id)
	}

	t.cancel()
	t.setStatus("paused")
	return nil
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

func Download(
	ctx context.Context,
	t *Torrent,
	tf *torrentfile.TorrentFile,
	peers []peer.Peer,
	peerID [20]byte,
) error {

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
		go worker(ctx, p, tf.InfoHash, peerID, workQueue, results)
	}

	var donePieces int
	for donePieces < len(tf.PieceHashes) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res := <-results:
			start, _ := tf.PieceBounds(res.index)
			if _, err := f.WriteAt(res.buf, int64(start)); err != nil {
				t.setStatus("error")
				return err
			}
			donePieces++
			t.pieceDone()
		}
	}
	t.setStatus("done")
	return nil

}
