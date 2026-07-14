package peer

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
)

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)
	buf[0] = byte(len(h.Pstr))
	copy(buf[1:], []byte(h.Pstr))
	copy(buf[28:48], h.InfoHash[:])
	copy(buf[48:], h.PeerID[:])

	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	lengthBuf := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, fmt.Errorf("cannot read pstrlen: %w", err)
	}
	pstrlen := int(lengthBuf[0])

	if pstrlen == 0 {
		return nil, fmt.Errorf("pstr cannot be 0")
	}
	buf := make([]byte, pstrlen+48)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read handshake body: %w", err)
	}

	h := &Handshake{Pstr: string(buf[0:pstrlen])}
	hashStart := pstrlen + 8
	copy(h.InfoHash[:], buf[hashStart:hashStart+20])
	copy(h.PeerID[:], buf[hashStart+20:hashStart+40])
	return h, nil
}

func Connect(p Peer, infoHash, peerID [20]byte) (net.Conn, error) {
	addr := net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}

	conn.SetDeadline(time.Now().Add(5 * time.Second))

	h := Handshake{Pstr: "BitTorent protocol", InfoHash: infoHash, PeerID: peerID}
	_, err = conn.Write(h.Serialize())
	if err != nil {
		conn.Close()
		return nil, err
	}

	res, err := ReadHandshake(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if res.InfoHash != infoHash {
		conn.Close()
		return nil, fmt.Errorf("InfoHash missmatch: peer wrong torrent")
	}
	conn.SetDeadline(time.Time{})
	return conn, nil
}
