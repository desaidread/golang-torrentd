package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"torrentd/internal/download"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: torrentd <file.torrent>")
	}

	m, err := download.NewManager()
	if err != nil {
		log.Fatal("manager:", err)
	}

	id, err := m.AddTorrent(os.Args[1])
	if err != nil {
		log.Fatal("add", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		fmt.Println("Остановка...")
		m.Shutdown()
		os.Exit(0)
	}()
	for {
		t, ok := m.Get(id)
		if !ok {
			break
		}

		done, total, status := t.Progress()
		pct := float64(done) / float64(total) * 100
		fmt.Printf("\r[%s] %d/%d кусков (%.1f%%)      ", status, done, total, pct)

		if status == "done" || status == "error" {
			break

		}
		time.Sleep(time.Second)
	}
}
