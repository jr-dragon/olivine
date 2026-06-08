package resp

import (
	"errors"
	"testing"
)

func TestSimpleString_Marshal(t *testing.T) {
	testcases := []struct {
		data   []byte
		expect []byte
	}{
		{
			data:   []byte("OK"),
			expect: []byte("+OK\r\n"),
		},
		{
			data:   []byte("PONG"),
			expect: []byte("+PONG\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(string(tc.data), func(t *testing.T) {
			s := SimpleString(tc.data)

			marshaled := s.Marshal()
			if string(marshaled) != string(tc.expect) {
				t.Errorf("expect '%s', got '%s'", tc.expect, marshaled)
			}
		})
	}
}

func BenchmarkSimpleString_bufferMarshal(b *testing.B) {
	s := SimpleString("OK")

	b.StartTimer()
	for b.Loop() {
		_ = s.bufferMarshal()
	}
	b.StopTimer()
}

func BenchmarkSimpleString_copyMarshal(b *testing.B) {
	s := SimpleString("OK")

	b.StartTimer()
	for b.Loop() {
		_ = s.copyMarshaler()
	}
	b.StopTimer()
}

func TestSimpleError_Marshal(t *testing.T) {
	err := NewSimpleError(errors.New("foobar"))
	marshaled := err.Marshal()
	expect := []byte("-ERR foobar\r\n")

	if string(marshaled) != string(expect) {
		t.Errorf("expect '%s', got '%s'", expect, marshaled)
	}
}

func TestInteger_Marshal(t *testing.T) {
	testcase := []struct {
		data   int
		expect []byte
	}{
		{
			data:   100,
			expect: []byte(":100\r\n"),
		},
		{
			data:   -100,
			expect: []byte(":-100\r\n"),
		},
	}

	for _, tc := range testcase {
		marshaled := Integer(tc.data).Marshal()

		if string(marshaled) != string(tc.expect) {
			t.Errorf("expect '%s', got '%s'", tc.expect, marshaled)
		}
	}
}

func TestBulkString_String(t *testing.T) {
	testcases := []struct {
		name   string
		data   *string
		expect string
	}{
		{
			name:   "null string",
			data:   nil,
			expect: "",
		},
		{
			name:   "empty string",
			data:   new(""),
			expect: "",
		},
		{
			name:   "normal string",
			data:   new("foobar"),
			expect: "foobar",
		},
		{
			name:   "binary string",
			data:   new("\x00\r\n\x00"),
			expect: "\x00\r\n\x00",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var str string
			if tc.data == nil {
				str = NewNullBulkString().String()
			} else {
				str = NewBulkString(*tc.data).String()
			}

			if str != tc.expect {
				t.Errorf("expect '%s', got '%s'", tc.expect, str)
			}
		})
	}
}

func TestBulkString_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		data   *string
		expect []byte
	}{
		{
			name:   "null string",
			data:   nil,
			expect: []byte("$-1\r\n"),
		},
		{
			name:   "empty string",
			data:   new(""),
			expect: []byte("$0\r\n\r\n"),
		},
		{
			name:   "normal string",
			data:   new("foobar"),
			expect: []byte("$6\r\nfoobar\r\n"),
		},
		{
			name:   "binary string",
			data:   new("\x00\r\n\x00"),
			expect: []byte("$4\r\n\x00\r\n\x00\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var marshaled []byte
			if tc.data == nil {
				marshaled = NewNullBulkString().Marshal()
			} else {
				marshaled = NewBulkString(*tc.data).Marshal()
			}

			if string(marshaled) != string(tc.expect) {
				t.Errorf("expect '%s', got '%s'", tc.expect, marshaled)
			}
		})
	}
}

func TestArray_Marshal(t *testing.T) {
	testcases := []struct {
		name   string
		data   Array
		expect []byte
	}{
		{
			name:   "null array",
			data:   NewNullArray(),
			expect: []byte("*-1\r\n"),
		},
		{
			name:   "empty array",
			data:   NewArray([]Value{}),
			expect: []byte("*0\r\n"),
		},
		{
			name: "command array",
			data: NewArray([]Value{
				NewBulkString("PING"),
				NewBulkString("message"),
			}),
			expect: []byte("*2\r\n$4\r\nPING\r\n$7\r\nmessage\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			marshaled := tc.data.Marshal()

			if string(marshaled) != string(tc.expect) {
				t.Errorf("expect '%s', got '%s'", tc.expect, marshaled)
			}
		})
	}
}
