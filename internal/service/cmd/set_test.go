package cmd

import (
	"context"
	"errors"
	"olivine/internal/repo"
	"olivine/pkg/resp"
	"slices"
	"testing"
)

func TestSet_Exec(t *testing.T) {
	testcases := []struct {
		name    string
		storage repo.Storage
		cmd     *resp.Command
		expect  any
	}{
		{
			name:    "missing argument",
			storage: &repo.StorageMock{},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
			})),
			expect: ErrValidation,
		},
		{
			name:    "too many arguments",
			storage: &repo.StorageMock{},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
			})),
			expect: ErrValidation,
		},
		{
			name: "set",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, k, v string) error { return nil },
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
			})),
			expect: []byte("+OK\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewSet(tc.storage)

			ret, err := cmd.Exec(t.Context(), tc.cmd)
			if experr, ok := tc.expect.(error); ok {
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				if !slices.Equal(tc.expect.([]byte), ret.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect.(string), ret.Marshal())
				}
			}
		})
	}
}
