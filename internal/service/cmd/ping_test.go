package cmd

import (
	"errors"
	"testing"

	"olivine/pkg/resp"
)

func TestPing_parse(t *testing.T) {
	testcases := []struct {
		name string
		args []resp.Value
		want error
	}{
		{
			name: "without message",
			args: []resp.Value{
				resp.NewBulkString("PING"),
			},
		},
		{
			name: "with message",
			args: []resp.Value{
				resp.NewBulkString("PING"),
				resp.NewBulkString("hello"),
			},
		},
		{
			name: "with too many messages",
			args: []resp.Value{
				resp.NewBulkString("PING"),
				resp.NewBulkString("hello"),
				resp.NewBulkString("world"),
			},
			want: ErrValidation,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			command := resp.NewTestCommand(resp.NewArray(tc.args))
			err := (&Ping{}).parse(command)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Validate() error = %v, want errors.Is(..., %v)", err, tc.want)
			}
		})
	}
}
