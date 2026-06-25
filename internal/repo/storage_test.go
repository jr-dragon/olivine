package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"olivine/internal/repo/object"
)

func TestMapStorage_SetString(t *testing.T) {
	t.Run("overwrites existing key without ttl", func(t *testing.T) {
		s := NewStorage()
		if err := s.Set(context.Background(), &setStringTestParam{key: "key", val: "old"}); err != nil {
			t.Fatalf("Set() old value error = %v", err)
		}

		if err := s.Set(context.Background(), &setStringTestParam{key: "key", val: "new"}); err != nil {
			t.Fatalf("Set() new value error = %v", err)
		}

		got := getString(t, s, "key")
		if got.String() != "new" {
			t.Errorf("stored value = %q, want %q", got.String(), "new")
		}
		if got.ExpiresAt() != nil {
			t.Errorf("stored expiration = %v, want nil", got.ExpiresAt())
		}
	})

	t.Run("keepttl on missing key stores without ttl", func(t *testing.T) {
		s := NewStorage()

		if err := s.Set(context.Background(), &setStringTestParam{key: "key", val: "value", keepTTL: true}); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		got := getString(t, s, "key")
		if got.String() != "value" {
			t.Errorf("stored value = %q, want %q", got.String(), "value")
		}
		if got.ExpiresAt() != nil {
			t.Errorf("stored expiration = %v, want nil", got.ExpiresAt())
		}
	})

	t.Run("expired key is missing for nx", func(t *testing.T) {
		s := NewStorage()
		expiredAt := time.Now().Add(-time.Second)
		if err := s.Set(context.Background(), &setStringTestParam{key: "key", val: "old", exp: &expiredAt}); err != nil {
			t.Fatalf("Set() old value error = %v", err)
		}

		if err := s.Set(context.Background(), &setStringTestParam{key: "key", val: "new", cond: CondNX}); err != nil {
			t.Fatalf("Set() new value error = %v", err)
		}

		got := getString(t, s, "key")
		if got.String() != "new" {
			t.Errorf("stored value = %q, want %q", got.String(), "new")
		}
		if got.ExpiresAt() != nil {
			t.Errorf("stored expiration = %v, want nil", got.ExpiresAt())
		}
	})
}

func TestMapStorage_Prune(t *testing.T) {
	testcases := []struct {
		name         string
		expiredCount int
	}{
		{
			name:         "below threshold",
			expiredCount: 2,
		},
		{
			name:         "at threshold",
			expiredCount: 3,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStorage().(*mapStorage)
			expiredAt := time.Now().Add(-time.Second)

			for i := range 10 {
				var expiresAt *time.Time
				if i < tc.expiredCount {
					expiresAt = &expiredAt
				}
				s.storage[fmt.Sprintf("key-%d", i)] = object.NewString(fmt.Sprintf("key-%d", i), "value", expiresAt)
			}

			if err := s.Prune(context.Background()); err != nil {
				t.Fatalf("Prune() error = %v", err)
			}

			if got, want := len(s.storage), 10-tc.expiredCount; got != want {
				t.Errorf("storage length = %d, want %d", got, want)
			}
		})
	}
}

type setStringTestParam struct {
	key       string
	val       string
	cond      Cond
	condValue string
	exp       *time.Time
	keepTTL   bool
	get       bool
	cur       *object.String
}

func (p *setStringTestParam) Obj() object.Object {
	return object.NewString(p.key, p.val, p.exp)
}

func (p *setStringTestParam) CondType() Cond {
	return p.cond
}

func (p *setStringTestParam) CondValue() string {
	return p.condValue
}

func (p *setStringTestParam) ExpiresAt() *time.Time {
	return p.exp
}

func (p *setStringTestParam) KeepTTL() bool {
	return p.keepTTL
}

func (p *setStringTestParam) GetCurrent() bool {
	return p.get
}

func (p *setStringTestParam) SetCurrent(cur *object.String) {
	p.cur = cur
}

func getString(t *testing.T, s Storage, key string) *object.String {
	t.Helper()

	got, err := s.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get(%q) error = %v", key, err)
	}
	str, ok := got.(*object.String)
	if !ok {
		t.Fatalf("Get(%q) = %T, want *object.String", key, got)
	}

	return str
}
