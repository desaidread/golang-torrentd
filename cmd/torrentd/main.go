package main

import (
	"fmt"
	"log"
	"os"
	"torrentd/internal/peer"
	"torrentd/internal/torrentfile"
	"torrentd/internal/tracker"
)

func main() {
	peerID, err := tracker.GeneratePeerID()
	if err != nil {
		log.Fatal("failed to generate peer id: ", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("usage: torrentd <file.torrent>")
	}
	path := os.Args[1]
	tf, err := torrentfile.Open(path)
	if err != nil {
		log.Fatal("failed to open torrentfile: ", err)
	}

	fmt.Println(tracker.Announce(tf, peerID, 6881))

	resp, err := tracker.Announce(tf, peerID, 6881)
	if err != nil {
		log.Fatal("failed to announce")
	}

	for _, v := range resp.Peers {
		conn, err := peer.Connect(v, tf.InfoHash, peerID)
		if err != nil {
			fmt.Println("Пусто: ", v.IP, err)
			continue
		}
		fmt.Println("Handshake выполнен успешно", v.IP)
		conn.Close()

	}
}
