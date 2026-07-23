package resp

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
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
		{
			name:   "string with carriage return",
			input:  "foo\rbar\r\n",
			expect: []byte("foo\rbar"),
		},
		{
			name:   "line larger than reader buffer",
			input:  strings.Repeat("a", 4096) + "\r\n",
			expect: []byte(strings.Repeat("a", 4096)),
		},
		{
			name:   "sentinel split across reader buffer",
			input:  strings.Repeat("a", 4095) + "\r\n",
			expect: []byte(strings.Repeat("a", 4095)),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			rd := NewReader(bytes.NewBufferString(tc.input))

			got, err := rd.readLine()
			if experr, ok := tc.expect.(error); ok {
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if !slices.Equal(tc.expect.([]byte), got) {
					t.Errorf("expect '%s', got '%s'", tc.expect, got)
				}
			}
		})
	}
}

func BenchmarkReader_readLine(b *testing.B) {
	testcases := []struct {
		name  string
		input string
	}{
		{
			name:  "short",
			input: "123\r\n",
		},
		{
			name:  "larger than reader buffer",
			input: strings.Repeat("1", 4096) + "\r\n",
		},
	}

	for _, tc := range testcases {
		b.Run(tc.name, func(b *testing.B) {
			source := strings.NewReader(tc.input)
			rd := NewReader(source)

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				source.Reset(tc.input)
				rd.rd.Reset(source)

				if _, err := rd.readLine(); err != nil {
					b.Fatal(err)
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
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
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
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if !slices.Equal(tc.expect.([]byte), v.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect.([]byte), v.Marshal())
				}
			}
		})
	}
}

func TestReader_ReadCommand(t *testing.T) {
	rd := NewReader(bytes.NewBufferString("*2\r\n$4\r\nPING\r\n$7\r\nmessage\r\n"))

	values, err := rd.ReadCommand()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"PING", "message"}
	if len(values) != len(want) {
		t.Fatalf("ReadCommand() returned %d values, want %d", len(values), len(want))
	}
	for i := range values {
		if got := values[i].String(); got != want[i] {
			t.Errorf("ReadCommand()[%d] = %q, want %q", i, got, want[i])
		}
	}
}
