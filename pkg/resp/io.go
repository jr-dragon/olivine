package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

var (
	ErrUnknownType = errors.New("unknown type")
)

type Reader struct {
	rd *bufio.Reader
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{rd: bufio.NewReader(rd)}
}

func (r *Reader) Read() (Value, error) {
	t, err := r.rd.ReadByte()
	if err != nil {
		return nil, err
	}

	switch t {
	case MAGIC_ARRAY:
		return r.readArray()
	case MAGIC_BULK_STRING:
		return r.readBulkString()
	default:
		return nil, fmt.Errorf("%w: %c", ErrUnknownType, t)
	}
}

func (r *Reader) Buffered() int {
	return r.rd.Buffered()
}

func (r *Reader) readArray() (Array, error) {
	sz, err := r.readInt()
	if err != nil {
		return Array{}, err
	}
	if sz < 0 { // Ref: https://redis.io/docs/latest/develop/reference/protocol-spec/#null-arrays
		return Array{null: true}, nil
	}

	arr := Array{}
	arr.data = make([]Value, 0, sz)

	for range sz {
		v, err := r.Read()
		if err != nil {
			return arr, err
		}

		arr.data = append(arr.data, v)
	}

	return arr, nil
}

func (r *Reader) readBulkString() (BulkString, error) {
	sz, err := r.readInt()
	if err != nil {
		return BulkString{}, err
	}
	if sz < 0 { // Ref: https://redis.io/docs/latest/develop/reference/protocol-spec/#null-bulk-strings
		return BulkString{null: true}, nil
	}

	buf := make([]byte, sz)
	if _, err := io.ReadFull(r.rd, buf); err != nil {
		return BulkString{}, err
	}

	// Drop tailing "\r\n" from reader.
	if cr, err := r.rd.ReadByte(); err != nil || cr != '\r' {
		return BulkString{}, errors.New("unexpected sentinel")
	}
	if nl, err := r.rd.ReadByte(); err != nil || nl != '\n' {
		return BulkString{}, errors.New("unexpected sentinel")
	}

	return BulkString{data: buf}, nil
}

func (r *Reader) readInt() (int, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(line))
}

func (r *Reader) readLine() ([]byte, error) {
	var fragments []byte

	for {
		data, err := r.rd.ReadSlice('\n')
		if errors.Is(err, bufio.ErrBufferFull) {
			fragments = append(fragments, data...)
			continue
		}
		if err != nil {
			return nil, err
		}

		if len(data) == 1 && data[0] == '\n' && len(fragments) > 0 && fragments[len(fragments)-1] == '\r' {
			return fragments[:len(fragments)-1], nil
		}

		if len(data) >= 2 && data[len(data)-2] == '\r' {
			data = data[:len(data)-2]
			if len(fragments) == 0 {
				return data, nil
			}

			return append(fragments, data...), nil
		}

		fragments = append(fragments, data...)
	}
}
