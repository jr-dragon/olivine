package repo

import (
	"context"
	"errors"
	"fmt"
	"olivine/internal/repo/object"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

//go:generate go tool moq -rm -out storage_mock.go . Storage
type Storage interface {
	Set(ctx context.Context, obj object.Object) error
	Get(ctx context.Context, k string) (object.Object, error)
}

func NewStorage() Storage {
	s := mapStorage{}
	s.storage = make(map[string]object.Object)

	return &s
}

type mapStorage struct {
	storage map[string]object.Object
	mu      sync.RWMutex
}

func (s *mapStorage) Set(_ context.Context, obj object.Object) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.storage[obj.Key()] = obj
	return nil
}

func (s *mapStorage) Get(_ context.Context, k string) (v object.Object, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
