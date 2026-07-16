package main

import (
	"fmt"
	"log"
	"os"
	"torrentd/internal/download"
	"torrentd/internal/torrentfile"
	"torrentd/internal/tracker"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: torrentd <file.torrent>")
	}
	path := os.Args[1]

	tf, err := torrentfile.Open(path)
	if err != nil {
		log.Fatal("open:", err)
	}

	peerID, err := tracker.GeneratePeerID()
	if err != nil {
		log.Fatal("peerID:", err)
	}

	resp, err := tracker.Announce(tf, peerID, 6881)
	if err != nil {
		log.Fatal("announce:", err)
	}

	fmt.Printf("Пиров: %d, кусков: %d\n", len(resp.Peers), len(tf.PieceHashes))

	buf, err := download.Download(tf, resp.Peers, peerID)
	if err != nil {
		log.Fatal("download:", err)
	}

	if err := os.WriteFile(tf.Name, buf, 0644); err != nil {
		log.Fatal("write:", err)
	}
	fmt.Println("Готово:", tf.Name)
}
