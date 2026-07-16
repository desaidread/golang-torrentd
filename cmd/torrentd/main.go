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

	for _, p := range resp.Peers {
		client, err := peer.NewClient(p, tf.InfoHash, peerID)
		if err != nil {
			fmt.Println("Пропуск", p.IP, err)
			continue
		}
		bf := client.Bitfield
		if len(bf) == 0 {
			fmt.Printf("%s: bitfield ПУСТОЙ (len=0)\n", p.IP)
		} else {
			fmt.Printf("%s: bitfield len=%d byte0=%08b hasPiece0=%v\n",
				p.IP, len(bf), bf[0], bf.HasPiece(0))
		}
		fmt.Println("Загрузка куска 0 у ", p.IP)
		buf, err := client.DownloadPiece(0, int(tf.PieceLength), tf.PieceHashes[0])
		client.Conn.Close()
		if err != nil {
			fmt.Println("Какая-то ошибка - кусок не скачался", err)
			continue
		}

		fmt.Println("Кусок 0 скачан, хэши совпали", len(buf))
		break

	}

}
