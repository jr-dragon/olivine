package resp

import (
	"bufio"
	"bytes"
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
	var buf bytes.Buffer

	for {
		data, err := r.rd.ReadBytes('\r')
		if err != nil {
			return nil, err
		}

		buf.Write(data)
		b, err := r.rd.ReadByte()
		if err != nil {
			return nil, err
		}
		if b == '\n' {
			return buf.Bytes()[:buf.Len()-1], nil
		}
		buf.WriteByte(b)
	}
}
