package torrentfile

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieecLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces       string
	PiecesLength int
	Length       int
	Name         string
}
