package cmd

import (
	"context"
	"errors"
	"olivine/internal/repo"
	"olivine/internal/repo/object"
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
				GetFunc: func(ctx context.Context, k string) (object.Object, error) { return nil, repo.ErrNotFound },
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("foo"),
			})),
			expect: []byte("$-1\r\n"),
		},
		{
			name: "wrong type",
			storage: &repo.StorageMock{
				GetFunc: func(ctx context.Context, k string) (object.Object, error) {
					return object.NewHash(k, nil), nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("foo"),
			})),
			expect: []byte("-ERR Operation against a key holding the wrong kind of value\r\n"),
		},
		{
			name: "found",
			storage: &repo.StorageMock{
				GetFunc: func(ctx context.Context, k string) (object.Object, error) {
					return object.NewString("foo", "bar", nil), nil
				},
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
				if !slices.Equal(tc.expect.([]byte), ret.Marshal()) {
					t.Errorf("expect '%s', got '%s'", tc.expect.([]byte), ret.Marshal())
				}
			}
		})
	}
}

func TestGet_parse(t *testing.T) {
	testcases := []struct {
		name string
		args []resp.Value
		want error
	}{
		{
			name: "without key",
			args: []resp.Value{
				resp.NewBulkString("GET"),
			},
			want: ErrValidation,
		},
		{
			name: "with key",
			args: []resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("key"),
			},
		},
		{
			name: "with too many keys",
			args: []resp.Value{
				resp.NewBulkString("GET"),
				resp.NewBulkString("key"),
				resp.NewBulkString("another-key"),
			},
			want: ErrValidation,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			command := resp.NewTestCommand(resp.NewArray(tc.args))
			err := (&Get{}).parse(command)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Validate() error = %v, want errors.Is(..., %v)", err, tc.want)
			}
		})
	}
}
