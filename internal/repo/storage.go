package repo

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"olivine/internal/repo/object"
)

var (
	ErrNotFound = errors.New("not found")
)

//go:generate go tool moq -rm -out storage_mock.go . Storage
type Storage interface {
	Set(ctx context.Context, param SetParam) error
	Get(ctx context.Context, k string) (object.Object, error)
	Prune(context.Context) error
}

type SetParam interface {
	Obj() object.Object
}

func NewStorage() Storage {
	s := mapStorage{}
	s.storage = make(map[string]object.Object)

	return &s
}

type mapStorage struct {
	storage map[string]object.Object
	mu      sync.Mutex
}

func (s *mapStorage) Set(_ context.Context, param SetParam) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[param.Obj().Key()] = param.Obj()
	return nil
}

func (s *mapStorage) Get(_ context.Context, k string) (v object.Object, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ok bool
	if v, ok = s.storage[k]; !ok {
		return nil, fmt.Errorf("%w: miss", ErrNotFound)
	}
	if v.ExpiresAt() != nil && time.Now().After(*v.ExpiresAt()) {
		delete(s.storage, k) // remove key from storage when expired
		return nil, fmt.Errorf("%w: expired", ErrNotFound)
	}

	return
}

func (s *mapStorage) Prune(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if stop := s.tryPrune(); stop {
				return nil
			}
		}
	}
}

func (s *mapStorage) tryPrune() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.storage) == 0 {
		return true
	}

	const sampleSize = 10

	sampled := 0
	expired := 0
	now := time.Now()

	for k, v := range s.storage {
		if sampled == sampleSize {
			break
		}

		sampled++
		if v.ExpiresAt() != nil && now.After(*v.ExpiresAt()) {
			delete(s.storage, k)
			expired++
		}
	}

	return expired*4 < sampled
}
