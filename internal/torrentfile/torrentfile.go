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
	PieceLength int64
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

	raw, err := d.Decode()
	if err != nil {
		return nil, fmt.Errorf("unable to decode file: %w", err)
	}

	start, end := d.GetInfoRaw()
	data = data[start:end]
	hash := sha1.Sum(data)

	top, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	announce, ok := top["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	info, ok := top["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	name, ok := info["name"].(string)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	length, ok := info["length"].(int64)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	pieceLength, ok := info["piece length"].(int64)
	if !ok {
		return nil, fmt.Errorf("cannot finish type assertion")
	}

	pieces, ok := info["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'pieces' field in info")
	}
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("corrupted file")
	}

	numPieces := len(pieces) / 20
	pieceHashes := make([][20]byte, numPieces)
	for i := 0; i < numPieces; i++ {
		copy(pieceHashes[i][:], pieces[i*20:(i+1)*20])
	}

	return &TorrentFile{
		InfoHash:    hash,
		Announce:    announce,
		Name:        name,
		Length:      length,
		PieceLength: pieceLength,
		PieceHashes: pieceHashes,
	}, nil
}

func (t *TorrentFile) PieceBounds(index int) (start, end int) {
	start = index * int(t.Length)
	end = start + int(t.Length)
	if end > int(t.Length) {
		end = int(t.Length)
	}
	return start, end
}
