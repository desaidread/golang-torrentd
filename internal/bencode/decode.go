package bencode

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	r         *bufio.Reader
	pos       int
	data      []byte
	infoStart int
	infoEnd   int
}

func NewDecoder(data []byte) *Decoder {
	r := bufio.NewReader(bytes.NewReader(data))
	return &Decoder{
		r:    r,
		data: data,
	}
}

func (d *Decoder) Decode() (any, error) {

	b, err := d.r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("unable to peek: %w", err)
	}

	switch {
	case b[0] == 'i':
		return d.decodeInt()
	case b[0] == 'l':
		return d.decodeList()
	case b[0] == 'd':
		return d.decodeDict()
	case b[0] >= '0' && b[0] <= '9':
		return d.decodeString()
	default:
		return nil, fmt.Errorf("unknown symbol: %v", b[0])
	}

}

func (d *Decoder) decodeInt() (int64, error) {
	d.readByte()

	data, err := d.readString('e')
	if err != nil {
		return 0, fmt.Errorf("decodeInt is unable to decode string: %w", err)
	}
	val, err := strconv.ParseInt(data[:len(data)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable convert string to int: %w", err)
	}
	return val, nil
}

func (d *Decoder) decodeString() (string, error) {
	lengthStr, err := d.readString(':')
	if err != nil {
		return "", fmt.Errorf("unable to read string length: %w", err)
	}
	lengthStr = lengthStr[:len(lengthStr)-1]

	length, err := strconv.Atoi(lengthStr)
	if length < 0 {
		return "", fmt.Errorf("invalid string length: %d", length)
	}
	if err != nil {
		return "", fmt.Errorf("unable to convert string length to int: %w", err)
	}
	buf := make([]byte, length)
	_, err = d.readFull(d.r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil

}

func (d *Decoder) decodeList() ([]any, error) {
	d.readByte()

	var list []any
	for {
		b, err := d.r.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("unable to peek in decodeList: %w", err)
		}
		if b[0] == 'e' {
			d.readByte()
			break
		}
		//Рекурентно декодируем следующий элемент
		item, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("unable to decode in list cycle: %w", err)
		}
		list = append(list, item)
	}
	return list, nil
}

func (d *Decoder) decodeDict() (map[string]any, error) {
	d.readByte()
	dict := make(map[string]any)

	for {
		b, err := d.r.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("unable to peek in decoderDict: %w", err)
		}

		if b[0] == 'e' {
			d.readByte()
			break
		}

		key, err := d.decodeString()
		if err != nil {
			return nil, fmt.Errorf("unable to decode key: %w", err)
		}
		var val any
		if key == "info" {
			d.infoStart = d.pos
			val, err = d.Decode()
			if err != nil {
				return nil, fmt.Errorf("unable to decode value: %w", err)
			}
			d.infoEnd = d.pos
		} else {
			val, err = d.Decode()
			if err != nil {
				return nil, fmt.Errorf("unable to decode value: %w", err)
			}
		}
		dict[key] = val

	}

	return dict, nil
}

// Обертки для изменения счётчика
func (d *Decoder) readByte() (byte, error) {
	byt, err := d.r.ReadByte()
	if err != nil {
		return 0, err
	}
	d.pos += 1
	return byt, err
}

func (d *Decoder) readString(delim byte) (string, error) {
	res, err := d.r.ReadString(delim)
	d.pos += len(res)
	return res, err

}

func (d *Decoder) readFull(r io.Reader, buf []byte) (int, error) {
	n, err := io.ReadFull(r, buf)
	d.pos += n
	return n, err
}

func (d *Decoder) GetInfoRaw() (InfoStart int, InfoEnd int) {
	return d.infoStart, d.infoEnd
}
