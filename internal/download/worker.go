package download

import (
	"time"
	"torrentd/internal/peer"
)

func worker(p peer.Peer,
	infoHash, peerID [20]byte,
	workQueue chan pieceWork,
	results chan pieceResult) {

	client, err := peer.NewClient(p, infoHash, peerID)
	if err != nil {
		return
	}
	defer client.Conn.Close()

	for pw := range workQueue {
		if !client.Bitfield.HasPiece(pw.index) {
			time.Sleep(50 * time.Millisecond)
			workQueue <- pw
			continue
		}
		buf, err := client.DownloadPiece(pw.index, pw.length, pw.hash)
		if err != nil {
			workQueue <- pw
			return
		}
		results <- pieceResult{index: pw.index, buf: buf}
	}
}
