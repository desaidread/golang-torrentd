package peer

import (
	"encoding/binary"
	"fmt"
	"net"
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
