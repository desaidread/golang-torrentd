package torrentfile

import (
	"crypto/sha1"
	"fmt"
	"os"
	"torrentd/internal/bencode"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int64
	Name        string
}

func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	return data, nil
}

func Open(path string) (*TorrentFile, error) {
	data, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	d := bencode.NewDecoder(data)

	_, err = d.Decode()
	if err != nil {
		return nil, fmt.Errorf("unable to decode file: %w", err)
	}

	start, end := d.GetInfoRaw()
	data = data[start:end]
	hash := sha1.Sum(data)
	return &TorrentFile{InfoHash: hash}, nil
}
