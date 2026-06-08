package resp

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"testing"
)

func TestReader_readLine(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		expect any
	}{
		{
			name:   "simple string line",
			input:  "foobar\r\n",
			expect: []byte("foobar"),
		},
		{
			name:   "without sentinel",
			input:  "invalid input",
			expect: io.EOF,
		},
		{
			name:   "string with newline",
			input:  "foo\nbar\r\n",
			expect: []byte("foo\nbar"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := NewReader(bytes.NewBufferString(tc.input))

			got, err := rd.readLine()
			if experr, ok := tc.expect.(error); ok {
				if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if string(tc.expect.([]byte)) != string(got) {
					t.Errorf("expect '%s', got '%s'", tc.expect, got)
				}
			}
		})
	}
}

func TestReader_readInt(t *testing.T) {
	testcase := []struct {
		name   string
		input  string
		expect any
	}{
		{
			name:   "positive number",
			input:  "100\r\n",
			expect: 100,
		},
		{
			name:   "negative number",
			input:  "-100\r\n",
			expect: -100,
		},
		{
			name:   "not number",
			input:  "invalid input\r\n",
			expect: strconv.ErrSyntax,
		},
		{
			name:   "floating point",
			input:  "100.10\r\n",
			expect: strconv.ErrSyntax,
		},
		{
			name:   "without sentinel",
			input:  "12345",
			expect: io.EOF,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			rd := NewReader(bytes.NewBufferString(tc.input))

			got, err := rd.readInt()
			if experr, ok := tc.expect.(error); ok {
				if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if tc.expect.(int) != got {
					t.Errorf("expect '%d', got '%d'", tc.expect, got)
				}
			}
		})
	}
}

func TestReader_Read(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		expect any
	}{
		{
			name:   "command",
			input:  "*1\r\n$4\r\nPING\r\n",
			expect: []byte("*1\r\n$4\r\nPING\r\n"),
		},
		{
			name:   "invalid type",
			input:  "^\r\n",
			expect: ErrUnknownType,
		},
		{
			name:   "without sentinel",
			input:  "$-1\r",
			expect: io.EOF,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := NewReader(bytes.NewBufferString(tc.input))

			v, err := rd.Read()
			if experr, ok := tc.expect.(error); ok {
				if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if string(tc.expect.([]byte)) != string(v.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect.([]byte), v.Marshal())
				}
			}
		})
	}
}
