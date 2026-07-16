package peer

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	peer     Peer
	infoHash [20]byte
	peerID   [20]byte
}

func NewClient(p Peer, infoHash, peerID [20]byte) (*Client, error) {
	conn, err := Connect(p, infoHash, peerID)
	if err != nil {
		return nil, err
	}
	msg, err := ReadMessage(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, fmt.Errorf("expected bitfield, got keep-alive")

	}
	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("not a bitfield received")
	}

	bitfield := Bitfield(msg.Payload)

	return &Client{Conn: conn,
		Choked:   true,
		Bitfield: bitfield,
		peer:     p,
		infoHash: infoHash,
		peerID:   peerID}, nil

}

func (c *Client) SendInterested() error {
	msg := Message{ID: MsgInterested}

	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := Message{ID: MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := Message{ID: MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendRequest(index, begin, length int) error {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	msg := Message{ID: MsgRequest, Payload: payload}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))

	msg := Message{ID: MsgHave, Payload: payload}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("Wrong message ID")
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("Message payload less than 8 byte")
	}
	parsedIndex := binary.BigEndian.Uint32(msg.Payload[0:4])
	if int(parsedIndex) != index {
		return 0, fmt.Errorf("wrong piece received")
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin is out of buffer range")
	}
	data := msg.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data is bigger than buffer")
	}
	copy(buf[begin:], data)
	return len(data), nil
}

func (c *Client) DownloadPiece(index, length int, hash [20]byte) ([]byte, error) {
	const (
		BlockSize  = 16384 //16 kb
		MaxBacklog = 5
	)
	buf := make([]byte, length)
	downloaded, requested, backlog := 0, 0, 0

	c.SendInterested()
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))

	for downloaded < length {
		if !c.Choked {
			for backlog < MaxBacklog && requested < length {
				blockSize := BlockSize
				if length-requested < blockSize { //Последний блок
					blockSize = length - requested
				}
				c.SendRequest(index, requested, blockSize)
				backlog++
				requested += blockSize
			}
		}
	}

	//Читаем сообщение
	msg, err := ReadMessage(c.Conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		return nil, nil
	}

	switch msg.ID {
	case MsgUnchoke:
		c.Choked = false
	case MsgChoke:
		c.Choked = true
	case MsgPiece:
		n, err := ParsePiece(index, buf, msg)
		if err != nil {
			return nil, err
		}
		downloaded += n
		backlog--
	}

	if sha1.Sum(buf) != hash {
		return nil, fmt.Errorf("piece %d failed hash check", index)
	}

	return buf, nil

}
