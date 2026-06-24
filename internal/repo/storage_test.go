package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"olivine/internal/repo/object"
)

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
