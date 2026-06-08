package resp

import (
	"bytes"
	"strconv"
)

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
	/*
	 * Simple String Format:
	 * |---------------------|---------|----------|
	 * | MAGIC_SIMPLE_STRING | Content | SENTINEL |
	 * |---------------------|---------|----------|
	 * |                  '+'| "hello" | "\r\n"   |
	 * |------------------------------------------|
	 */
	var buf bytes.Buffer

	buf.WriteRune(MAGIC_SIMPLE_STRING)
	buf.Write(v)
	buf.WriteString(SENTINEL)

	return buf.Bytes()
}

type SimpleError struct {
	err error
}

func NewSimpleError(err error) SimpleError {
	return SimpleError{err: err}
}

func (v SimpleError) Marshal() []byte {
	var buf bytes.Buffer

	buf.WriteRune(MAGIC_SIMPLE_ERROR)
	buf.WriteString("ERR ")
	buf.WriteString(v.err.Error())
	buf.WriteString(SENTINEL)

	return buf.Bytes()
}

type Integer int

func (v Integer) Marshal() []byte {
	var buf bytes.Buffer

	buf.WriteRune(MAGIC_INTEGER)
	buf.WriteString(strconv.Itoa(int(v)))
	buf.WriteString(SENTINEL)

	return buf.Bytes()
}

type BulkString struct {
	null bool
	data []byte
}

func NewNullBulkString() BulkString {
	return BulkString{null: true}
}

func NewBulkString(s string) BulkString {
	return BulkString{data: []byte(s)}
}

func (v BulkString) Marshal() []byte {
	/*
	 * Null Bulk String Format:
	 * |-------------------|--------|----------|
	 * | MAGIC_BULK_STRING | LENGTH | SENTINEL |
	 * |-------------------|--------|----------|
	 * |               '-' | "-1"   | "\r\n"   |
	 * |-------------------|--------|----------|
	 */
	if v.null {
		return []byte("$-1\r\n")
	}

	/*
	 * Non Null Bulk String Format:
	 * |-------------------|--------|---------|----------|
	 * | MAGIC_BULK_STRING | Length | Content | SENTINEL |
	 * |-------------------|--------|---------|----------|
	 * |               '$' |    "5" | "hello" | "\r\n"   |
	 * |-------------------|--------|---------|----------|
	 */
	var buf bytes.Buffer
	buf.WriteRune(MAGIC_BULK_STRING)
	buf.WriteString(strconv.Itoa(len(v.data)))
	buf.WriteString(SENTINEL)
	buf.Write(v.data)
	buf.WriteString(SENTINEL)
	return buf.Bytes()
}

type Array struct {
	null bool
	data []Value
}

func (v Array) Marshal() []byte {
	/*
	 * Null Array Format:
	 * |-------------|--------|----------|
	 * | MAGIC_ARRAY | LENGTH | SENTINEL |
	 * |-------------|--------|----------|
	 * |         '*' | "-1"   | "\r\n"   |
	 * |-------------|--------|----------|
	 */
	if v.null {
		return []byte("*-1\r\n")
	}

	/*
	 * Array Format:
	 * |-------------|--------|----------|-----------------------|
	 * | MAGIC_ARRAY | LENGTH | SENTINEL | Marshaled Element [0] |
	 * |-------------|--------|----------|-----------------------|
	 * |         '*' |    "1" | "\r\n"   | "$4\r\nPING\r\n"      |
	 * |-------------|--------|----------|-----------------------|
	 */
	var buf bytes.Buffer
	buf.WriteRune(MAGIC_ARRAY)
	buf.WriteString(strconv.Itoa(len(v.data)))
	buf.WriteString(SENTINEL)
	for _, v := range v.data {
		buf.Write(v.Marshal())
	}
	return buf.Bytes()
}
