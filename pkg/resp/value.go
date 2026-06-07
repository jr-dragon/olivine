package resp

import "bytes"

const (
	MAGIC_SIMPLE_STRING = '+'
	MAGIC_SIMPLE_ERROR  = '-'
	MAGIC_INTEGER       = ':'
	MAGIC_BULK_STRING   = '$'
	MAGIC_ARRAY         = '*'
	SENTINEL            = "\r\n"
)

type Value interface {
	Marshal() []byte
}

type SimpleString []byte

func (v SimpleString) Marshal() []byte {
	var buf bytes.Buffer

	buf.WriteRune(MAGIC_SIMPLE_STRING)
	buf.Write(v)
	buf.WriteString(SENTINEL)

	return buf.Bytes()
}
