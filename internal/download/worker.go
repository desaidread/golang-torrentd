package download

import (
	"context"
	"torrentd/internal/peer"
)

func worker(
	ctx context.Context,
	p peer.Peer,
	infoHash, peerID [20]byte,
	workQueue chan pieceWork,
	results chan pieceResult) {

	client, err := peer.NewClient(p, infoHash, peerID)
	if err != nil {
		return
	}
	defer client.Conn.Close()

	for {

		select {
		case <-ctx.Done():
			return
		case pw := <-workQueue:
			if !client.Bitfield.HasPiece(pw.index) {
				select {
				case workQueue <- pw:
				case <-ctx.Done():
					return
				}
				continue
			}
			buf, err := client.DownloadPiece(pw.index, pw.length, pw.hash)
			if err != nil {
				select {
				case workQueue <- pw:
				case <-ctx.Done():
					return
				}
				return
			}
			select {
			case results <- pieceResult{index: pw.index, buf: buf}:
			case <-ctx.Done():
			}

		}
	}
}
