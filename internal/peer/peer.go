package peer

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func ParsePeers(peerBytes []byte) ([]Peer, error) {

	const peerSize = 6
	if len(peerBytes)%peerSize != 0 {
		return nil, fmt.Errorf("Received malformed peers data")
	}
	numPeers := len(peerBytes) / peerSize
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offSet := i * peerSize
		peers[i].IP = net.IP(peerBytes[offSet : offSet+4])

		peers[i].Port = binary.BigEndian.Uint16(peerBytes[offSet+4 : offSet+6])

	}

	return peers, nil
}
