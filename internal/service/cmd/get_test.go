package cmd

import (
	"context"
	"errors"
	"olivine/internal/repo"
	"olivine/pkg/resp"
	"slices"
	"testing"
)

func TestGet_Exec(t *testing.T) {
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
				resp.NewBulkString("GET"),
			})),
			expect: ErrValidation,
		},
		{
			name:    "too many arguments",
			storage: &repo.StorageMock{},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
			})),
			expect: ErrValidation,
		},
		{
			name: "not found",
			storage: &repo.StorageMock{
				GetFunc: func(ctx context.Context, k string) (string, error) { return "", repo.ErrNotFound },
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("foo"),
			})),
			expect: []byte("$-1\r\n"),
		},
		{
			name: "found",
			storage: &repo.StorageMock{
				GetFunc: func(ctx context.Context, k string) (string, error) { return "bar", nil },
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("foo"),
			})),
			expect: []byte("$3\r\nbar\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewGet(tc.storage)

			ret, err := cmd.Exec(t.Context(), tc.cmd)
			if experr, ok := tc.expect.(error); ok {
				if err == nil {
					t.Errorf("expect '%s', got nil", experr.Error())
				} else if !errors.Is(err, experr) {
					t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
				}
			} else {
				// if tc.expect.(string) != string(ret.Marshal()) {
				if !slices.Equal(tc.expect.([]byte), ret.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect.(string), ret.Marshal())
				}
			}
		})
	}
}
