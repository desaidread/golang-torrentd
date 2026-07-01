package bencode

import (
	"bufio"
)

type Decoder struct {
	r *bufio.Reader
}
