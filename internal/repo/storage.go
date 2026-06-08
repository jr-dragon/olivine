package repo

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type Storage interface {
	Set(ctx context.Context, k, v string) error
	Get(ctx context.Context, k string) (string, error)
}

func NewStorage() Storage {
	s := mapStorage{}
	s.storage = make(map[string]string)

	return &s
}

type mapStorage struct {
	storage map[string]string
	mu      sync.RWMutex
}

func (s *mapStorage) Set(_ context.Context, k, v string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[k] = v
	return nil
}

func (s *mapStorage) Get(_ context.Context, k string) (v string, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ok bool
	if v, ok = s.storage[k]; !ok {
		return "", ErrNotFound
	}

	return
}
