package data

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewConfigFromBytes(t *testing.T) {
	testcases := []struct {
		name string
		data []byte
		want any
	}{
		{
			name: "normal case",
			data: []byte("appendonly yes\nappendfsync no\n"),
			want: &Config{AOFEnabled: true, AOFFsync: AOFFsyncNo},
		},
		{
			name: "without tailing newline",
			data: []byte("appendonly yes\nappendfsync no\n"),
			want: &Config{AOFEnabled: true, AOFFsync: AOFFsyncNo},
		},
		{
			name: "unexpect div",
			data: []byte("appendonly\tyes\nappendfsync\tno\n"),
			want: &Config{AOFEnabled: true, AOFFsync: AOFFsyncNo},
		},
		{
			name: "invalid command: missing command argument",
			data: []byte("appendonly"),
			want: ErrCommand,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfigFromBytes(tt.data)

			if experr, ok := tt.want.(error); ok {
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if diff := cmp.Diff(got, tt.want); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}
