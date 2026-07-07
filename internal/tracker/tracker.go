package tracker

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"torrentd/internal/bencode"
	"torrentd/internal/peer"
	"torrentd/internal/torrentfile"
)

type Response struct {
	Interval int64
	Peers    []peer.Peer
}

func BuildURL(tf *torrentfile.TorrentFile, peerID [20]byte, port int) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", fmt.Errorf("cannot parse url from torrentfile.Announce: %w", err)
	}

	params := url.Values{}
	params.Set("info_hash", string(tf.InfoHash[:]))
	params.Set("peer_id", string(peerID[:]))
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("left", strconv.FormatInt(tf.Length, 10))
	params.Set("compact", "1")

	base.RawQuery = params.Encode()
	finalUrl := base.String()

	return finalUrl, nil
}

func GeneratePeerID() ([20]byte, error) {
	var peerID [20]byte
	var empty [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return empty, fmt.Errorf("unable to generate peerID: %w", err)
	}

	return peerID, nil
}

func Announce(tf *torrentfile.TorrentFile, peerID [20]byte, port int) (*Response, error) {

	URL, err := BuildURL(tf, peerID, port)
	if err != nil {
		return nil, err
	}
	r, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("Unable to make get request: %w", err)
	}

	defer r.Body.Close()

	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot extract data from reposponse.body: %w", err)
	}

	dataRaw, err := bencode.NewDecoder(responseData).Decode()
	if err != nil {
		return nil, err
	}
	top, ok := dataRaw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unable to conver response data")
	}

	interval, ok := top["interval"].(int64)
	if !ok {
		return nil, fmt.Errorf("nnable to conver response data")
	}

	peers, ok := top["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("unable to conver response data")
	}

	peersList, err := peer.ParsePeers([]byte(peers))
	if err != nil {
		return nil, fmt.Errorf("Unable to parse peers: %w", err)
	}

	return &Response{
		Interval: interval,
		Peers:    peersList,
	}, nil
}
