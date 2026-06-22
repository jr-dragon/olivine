package cmd

import (
	"context"
	"errors"
	"fmt"
	"olivine/internal/repo"
	"olivine/internal/repo/object"
	"olivine/pkg/resp"
	"slices"
	"testing"
	"testing/synctest"
	"time"
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
				SetFunc: func(ctx context.Context, obj object.Object) error { return nil },
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
			})),
			expect: []byte("+OK\r\n"),
		},
		{
			name: "set with ex",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, obj object.Object) error {
					if obj.ExpiresAt() == nil || !obj.ExpiresAt().Equal(time.Now().Add(time.Second*10)) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", time.Now().Add(time.Second*10), obj.ExpiresAt())
					}
					return nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
				resp.NewBulkString("EX"),
				resp.NewBulkString("10"),
			})),
			expect: []byte("+OK\r\n"),
		},
		{
			name: "set with px",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, obj object.Object) error {
					if obj.ExpiresAt() == nil || !obj.ExpiresAt().Equal(time.Now().Add(time.Millisecond*11)) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", time.Now().Add(time.Millisecond*11), obj.ExpiresAt())
					}
					return nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
				resp.NewBulkString("PX"),
				resp.NewBulkString("11"),
			})),
			expect: []byte("+OK\r\n"),
		},
		{
			name: "set with exat",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, obj object.Object) error {
					expected := time.Unix(1_700_000_000, 0)
					if obj.ExpiresAt() == nil || !obj.ExpiresAt().Equal(expected) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", expected, obj.ExpiresAt())
					}
					return nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
				resp.NewBulkString("EXAT"),
				resp.NewBulkString("1700000000"),
			})),
			expect: []byte("+OK\r\n"),
		},
		{
			name: "set with pxat",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, obj object.Object) error {
					expected := time.UnixMilli(1_700_000_000_123)
					if obj.ExpiresAt() == nil || !obj.ExpiresAt().Equal(expected) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", expected, obj.ExpiresAt())
					}
					return nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
				resp.NewBulkString("PXAT"),
				resp.NewBulkString("1700000000123"),
			})),
			expect: []byte("+OK\r\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				cmd := NewSet(tc.storage)

				ret, err := cmd.Exec(t.Context(), tc.cmd)
				if experr, ok := tc.expect.(error); ok {
					if err == nil {
						t.Errorf("expect '%s', got nil", experr.Error())
					} else if !errors.Is(err, experr) {
						t.Errorf("expect '%s', got '%s'", experr.Error(), err.Error())
					}
				} else if err != nil {
					t.Errorf("expect '%s', got '%s'", tc.expect.([]byte), err.Error())
				} else {
					if !slices.Equal(tc.expect.([]byte), ret.Marshal()) {
						t.Errorf("expect '%s', got '%s'", tc.expect.([]byte), ret.Marshal())
					}
				}
			})
		})
	}
}
