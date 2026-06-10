package resp

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func TestReadCommand(t *testing.T) {
	testcases := []struct {
		name   string
		input  string
		expect any
	}{
		{
			name:   "not array",
			input:  "$4\r\nPING\r\n",
			expect: ErrProtocol,
		},
		{
			name:   "null array",
			input:  "*-1\r\n",
			expect: ErrProtocol,
		},
		{
			name:   "empty array",
			input:  "*0\r\n",
			expect: ErrProtocol,
		},
		{
			name:   "command is not bulk string",
			input:  "*1\r\n*0\r\n",
			expect: ErrProtocol,
		},
		{
			name:   "arguments is not bulk string",
			input:  "*2\r\n$4\r\nPING\r\n*-1\r\n",
			expect: ErrProtocol,
		},
		{
			name:   "valid command",
			input:  "*2\r\n$4\r\nPING\r\n$7\r\nmessage\r\n",
			expect: []byte("*2\r\n$4\r\nPING\r\n$7\r\nmessage\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := ReadCommand(NewReader(bytes.NewBufferString(tc.input)))

			if experr, ok := tc.expect.(error); ok {
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if !slices.Equal(tc.expect.([]byte), cmd.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect, cmd.Marshal())
				}
			}
		})
	}
}
