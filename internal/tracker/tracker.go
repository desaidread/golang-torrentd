package tracker

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"torrentd/internal/bencode"
	"torrentd/internal/torrentfile"
)

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
		return empty, fmt.Errorf("unnable to generate peerID: %w", err)
	}

	return peerID, nil
}

func Announce(tf *torrentfile.TorrentFile, peerID [20]byte, port int) (any, error) {

	URL, err := BuildURL(tf, peerID, port)
	if err != nil {
		return nil, err
	}
	r, err := http.Get(URL)
	if err != nil {
		return nil, fmt.Errorf("Unablle to make get request: %w", err)
	}

	defer r.Body.Close()

	responseData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot extract data from reposponse.body: %w", err)
	}

	data, err := bencode.NewDecoder(responseData).Decode()
	if err != nil {
		return nil, err
	}
	return data, nil

}
