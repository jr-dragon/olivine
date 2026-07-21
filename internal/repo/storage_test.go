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

func TestMapStorage_TryPrune(t *testing.T) {
	const sampleSize = 10

	t.Run("empty storage", func(t *testing.T) {
		s := NewStorage().(*mapStorage)

		if got := s.tryPrune(); !got {
			t.Errorf("tryPrune() = %t, want true", got)
		}
	})

	testcases := []struct {
		name         string
		expiredCount int
		wantStop     bool
	}{
		{
			name:         "below threshold",
			expiredCount: 2,
			wantStop:     true,
		},
		{
			name:         "at threshold",
			expiredCount: 3,
			wantStop:     false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStorage().(*mapStorage)
			expiredAt := time.Now().Add(-time.Second)
			counts := [lockStripeCount]int{}
			filled := 0

			for i := 0; filled < lockStripeCount; i++ {
				key := fmt.Sprintf("key-%d", i)
				slot := s.stripe(key)
				if counts[slot] == sampleSize {
					continue
				}

				var expiresAt *time.Time
				if counts[slot] < tc.expiredCount {
					expiresAt = &expiredAt
				}
				s.storage[slot][key] = object.NewString(key, "value", expiresAt)
				counts[slot]++
				if counts[slot] == sampleSize {
					filled++
				}
			}

			if got := s.tryPrune(); got != tc.wantStop {
				t.Errorf("tryPrune() = %t, want %t", got, tc.wantStop)
			}

			if got, want := storageLength(s), lockStripeCount*sampleSize-tc.expiredCount; got != want {
				t.Errorf("storage length = %d, want %d", got, want)
			}
		})
	}
}

func TestMapStorage_TryPruneConcurrentSet(t *testing.T) {
	s := NewStorage().(*mapStorage)
	ctx := context.Background()
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := range 1_000 {
			key := fmt.Sprintf("key-%d", i%100)
			if err := s.Set(ctx, &setStringTestParam{key: key, val: "value"}); err != nil {
				t.Errorf("Set(%q) error = %v", key, err)
				return
			}
		}
	}()

	for range 1_000 {
		s.tryPrune()
	}
	<-done
}

func storageLength(s *mapStorage) int {
	var length int
	for _, stripe := range s.storage {
		length += len(stripe)
	}

	return length
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
