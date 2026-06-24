package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	"olivine/internal/repo"
	"olivine/pkg/resp"
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
				SetFunc: func(ctx context.Context, p repo.SetParam) error { return nil },
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
				SetFunc: func(ctx context.Context, p repo.SetParam) error {
					if p.Obj().ExpiresAt() == nil || !p.Obj().ExpiresAt().Equal(time.Now().Add(time.Second*10)) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", time.Now().Add(time.Second*10), p.Obj().ExpiresAt())
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
				SetFunc: func(ctx context.Context, p repo.SetParam) error {
					if p.Obj().ExpiresAt() == nil || !p.Obj().ExpiresAt().Equal(time.Now().Add(time.Millisecond*10)) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", time.Now().Add(time.Millisecond*10), p.Obj().ExpiresAt())
					}
					return nil
				},
			},
			cmd: resp.NewTestCommand(resp.NewArray([]resp.Value{
				resp.NewBulkString("SET"),
				resp.NewBulkString("foo"),
				resp.NewBulkString("bar"),
				resp.NewBulkString("PX"),
				resp.NewBulkString("10"),
			})),
			expect: []byte("+OK\r\n"),
		},
		{
			name: "set with exat",
			storage: &repo.StorageMock{
				SetFunc: func(ctx context.Context, p repo.SetParam) error {
					expected := time.Unix(1_700_000_000, 0)
					if p.Obj().ExpiresAt() == nil || !p.Obj().ExpiresAt().Equal(expected) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", expected, p.Obj().ExpiresAt())
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
				SetFunc: func(ctx context.Context, p repo.SetParam) error {
					expected := time.UnixMilli(1_700_000_000_123)
					if p.Obj().ExpiresAt() == nil || !p.Obj().ExpiresAt().Equal(expected) {
						return fmt.Errorf("unexpected expire time: expected '%v', got '%v'", expected, p.Obj().ExpiresAt())
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

func TestSet_parse(t *testing.T) {
	testcases := []struct {
		name         string
		args         []resp.Value
		wantCond     parseSetCond
		wantGet      bool
		wantExp      *time.Time
		wantExpAfter time.Duration
		wantKeepTTL  bool
		wantErr      error
	}{
		{
			name: "without options",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
			},
		},
		{
			name: "NX with GET and EX",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
				resp.NewBulkString("NX"), resp.NewBulkString("GET"), resp.NewBulkString("EX"), resp.NewBulkString("10"),
			},
			wantCond:     parseSetCond{Typ: nx},
			wantGet:      true,
			wantExpAfter: 10 * time.Second,
		},
		{
			name: "conditional value with PXAT",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
				resp.NewBulkString("IFEQ"), resp.NewBulkString("expected"), resp.NewBulkString("PXAT"), resp.NewBulkString("1700000000123"),
			},
			wantCond: parseSetCond{Typ: ifeq, Val: "expected"},
			wantExp:  new(time.UnixMilli(1_700_000_000_123)),
		},
		{
			name: "KEEPTTL",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"), resp.NewBulkString("KEEPTTL"),
			},
			wantKeepTTL: true,
		},
		{
			name: "missing value",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "condition without comparison value",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"), resp.NewBulkString("IFNE"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "expiration without value",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"), resp.NewBulkString("EX"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "option after expiration",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
				resp.NewBulkString("EX"), resp.NewBulkString("10"), resp.NewBulkString("GET"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "duplicate condition",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
				resp.NewBulkString("NX"), resp.NewBulkString("XX"),
			},
			wantErr: ErrValidation,
		},
		{
			name: "duplicate GET",
			args: []resp.Value{
				resp.NewBulkString("SET"), resp.NewBulkString("key"), resp.NewBulkString("value"),
				resp.NewBulkString("GET"), resp.NewBulkString("GET"),
			},
			wantErr: ErrValidation,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				command := resp.NewTestCommand(resp.NewArray(tc.args))
				got, err := (&Set{}).parse(command)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("parse() error = %v, want errors.Is(..., %v)", err, tc.wantErr)
				}
				if err != nil {
					return
				}
				if got.Cond != tc.wantCond {
					t.Errorf("parse() condition = %#v, want %#v", got.Cond, tc.wantCond)
				}
				if got.Get != tc.wantGet {
					t.Errorf("parse() GET = %t, want %t", got.Get, tc.wantGet)
				}
				if tc.wantKeepTTL {
					if got.Exp == nil || !got.Exp.IsZero() {
						t.Errorf("parse() expiration = %v, want non-nil zero time", got.Exp)
					}
					return
				}
				wantExp := tc.wantExp
				if tc.wantExpAfter != 0 {
					wantExp = new(time.Now().Add(tc.wantExpAfter))
				}
				if got.Exp == nil || wantExp == nil {
					if got.Exp != wantExp {
						t.Errorf("parse() expiration = %v, want %v", got.Exp, wantExp)
					}
				} else if !got.Exp.Equal(*wantExp) {
					t.Errorf("parse() expiration = %v, want %v", got.Exp, wantExp)
				}
			})
		})
	}
}
