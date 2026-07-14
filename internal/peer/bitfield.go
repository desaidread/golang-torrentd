package peer

type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	bitOffset := index % 8
	bit := bf[byteIndex] >> (7 - bitOffset) & 1
	return bit != 0
}
