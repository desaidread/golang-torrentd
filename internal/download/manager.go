package download

import (
	"fmt"
	"torrentd/internal/peer"
	"torrentd/internal/torrentfile"
)

func Download(tf *torrentfile.TorrentFile, peers []peer.Peer, peerID [20]byte) ([]byte, error) {
	workQueue := make(chan pieceWork, len(tf.PieceHashes))
	results := make(chan pieceResult)

	for index, hash := range tf.PieceHashes {
		start, end := tf.PieceBounds(index)
		length := end - start
		workQueue <- pieceWork{index: index, hash: hash, length: length}
	}

	for _, p := range peers {
		go worker(p, tf.InfoHash, peerID, workQueue, results)
	}

	var donePieces int
	buf := make([]byte, tf.Length)
	for donePieces < len(tf.PieceHashes) {
		res := <-results
		start, _ := tf.PieceBounds(res.index)
		copy(buf[start:], res.buf)
		donePieces++
		precent := float64(donePieces) / float64(len(tf.PieceHashes)) * 100
		fmt.Printf("(%.2f%%) кусок %d готов\n", precent, res.index)

	}
	close(workQueue)
	return buf, nil
}
