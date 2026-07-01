package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r *bufio.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode() (interface{}, error) {

	b, err := d.r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("unable to peek: %v", err)
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
		return nil, fmt.Errorf("Unknown symbol: %v", b[0])
	}

}

func (d *Decoder) decodeInt() (int64, error) {
	d.r.ReadByte()

	data, err := d.r.ReadString('e')
	if err != nil {
		return 0, fmt.Errorf("decodeInt is unnable to decode string: %v", err)
	}
	val, err := strconv.ParseInt(data[:len(data)-1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unable convert sttring to int: %v ", err)
	}
	return val, nil
}

func (d *Decoder) decodeString() (string, error) {
	lengthStr, err := d.r.ReadString(':')
	if err != nil {
		return "", fmt.Errorf("unable to read string legth: %v", err)
	}
	lengthStr = lengthStr[:len(lengthStr)-1]

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", fmt.Errorf("unable to convert string length to int: %v", err)
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(d.r, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil

}

func (d *Decoder) decodeList() ([]interface{}, error) {
	d.r.ReadByte()

	var list []interface{}
	for {
		b, err := d.r.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("unable to peek in decodeList: %v", err)
		}
		if b[0] == 'e' {
			d.r.ReadByte()
			break
		}
		//Рекурентно декодируем следующий элемент
		item, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("unable to decode in list cycle: %v", err)
		}
		list = append(list, item)
	}
	return list, nil
}

func (d *Decoder) decodeDict() (map[string]interface{}, error) {
	d.r.ReadByte()
	dict := make(map[string]interface{})

	for {
		b, err := d.r.Peek(1)
		if err != nil {
			return nil, fmt.Errorf("Unable to peek in decoderDict: %v", err)
		}

		if b[0] == 'e' {
			d.r.ReadByte()
			break
		}

		key, err := d.decodeString()
		if err != nil {
			return nil, fmt.Errorf("Unable to decode key: %v", err)
		}

		val, err := d.Decode()
		if err != nil {
			return nil, fmt.Errorf("Unable to decode value: %v", err)
		}
		dict[key] = val

	}

	return dict, nil
}
