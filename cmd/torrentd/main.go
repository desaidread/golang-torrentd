package main

import (
	"fmt"
	"log"
	"os"
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

}
