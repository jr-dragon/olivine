package repo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"olivine/internal/repo/object"

	"github.com/zeebo/xxh3"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrTypeMismatch = errors.New("type mismatch")
	ErrCondMismatch = errors.New("condition mismatch")
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

type Cond int

const (
	CondNX    Cond = iota + 1 // set value only not exists
	CondXX                    // set value only exists
	CondIFEQ                  // set only value == cond.Val
	CondIFNE                  // set only value != cond.Val
	CondIFDEQ                 // set only XXH3(value) == cond.Val
	CondIFDNE                 // set only XXH3(value) != cond.Val
)

type SetStringParam interface {
	SetParam

	CondType() Cond
	CondValue() string
	ExpiresAt() *time.Time
	KeepTTL() bool

	GetCurrent() bool
	SetCurrent(*object.String)
}

func NewStorage() Storage {
	s := mapStorage{}
	for i := range lockStripeCount {
		s.storage[i] = make(map[string]object.Object)
	}

	return &s
}

const lockStripeCount = 16

type mapStorage struct {
	storage [lockStripeCount]map[string]object.Object
	stripes [lockStripeCount]sync.Mutex
}

func (s *mapStorage) Set(_ context.Context, param SetParam) error {
	if strparam, ok := param.(SetStringParam); ok {
		return s.setString(strparam)
	}

	return errors.New("unimplemented")
}

func (s *mapStorage) setString(param SetStringParam) error {
	obj := param.Obj()

	slot := s.stripe(obj.Key())
	s.stripes[slot].Lock()
	defer s.stripes[slot].Unlock()

	cur, exists := s.storage[slot][obj.Key()]
	if cur != nil && cur.Expired() {
		delete(s.storage[slot], obj.Key())
		cur = nil
		exists = false
	}

	if param.GetCurrent() || param.KeepTTL() {
		if !exists {
			param.SetCurrent(nil)
		} else {
			if curstr, ok := cur.(*object.String); !ok {
				return ErrTypeMismatch
			} else {
				param.SetCurrent(curstr)
			}
		}
	}
	if err := s.checkStringCond(param, cur, exists); err != nil {
		return fmt.Errorf("%w: %w", ErrCondMismatch, err)
	}
	if param.KeepTTL() && cur != nil {
		obj.SetExpiresAt(cur.ExpiresAt())
	}

	s.storage[slot][obj.Key()] = obj

	return nil
}

func (s *mapStorage) checkStringCond(param SetStringParam, current object.Object, exists bool) error {
	switch param.CondType() {
	case 0:
		return nil
	case CondNX:
		if exists {
			return errors.New("data found")
		}
	case CondXX:
		if !exists {
			return errors.New("data not found")
		}
	case CondIFEQ:
		if !exists {
			return errors.New("data not found")
		}
		if str, ok := current.(*object.String); !ok {
			return ErrTypeMismatch
		} else if !str.Equals(param.CondValue()) {
			return errors.New("data mismatch")
		}
	case CondIFNE:
		if !exists {
			return nil // ignore when current doesn't exist
		}
		if str, ok := current.(*object.String); !ok {
			return ErrTypeMismatch
		} else if str.Equals(param.CondValue()) {
			return errors.New("data match")
		}
	case CondIFDEQ:
		if !exists {
			return errors.New("data not found")
		}
		if str, ok := current.(*object.String); !ok {
			return ErrTypeMismatch
		} else if !str.EqualsDigest(param.CondValue()) {
			return errors.New("data match")
		}
	case CondIFDNE:
		if !exists {
			return nil // ignore when current doesn't exist
		}
		if str, ok := current.(*object.String); !ok {
			return ErrTypeMismatch
		} else if str.EqualsDigest(param.CondValue()) {
			return errors.New("data match")
		}
	default:
		panic(fmt.Sprintf("unknow condition type: %d", param.CondType()))
	}

	return nil
}

func (s *mapStorage) Get(_ context.Context, k string) (v object.Object, err error) {
	slot := s.stripe(k)
	s.stripes[slot].Lock()
	defer s.stripes[slot].Unlock()

	var ok bool
	if v, ok = s.storage[slot][k]; !ok {
		return nil, fmt.Errorf("%w: miss", ErrNotFound)
	}
	if v.ExpiresAt() != nil && time.Now().After(*v.ExpiresAt()) {
		delete(s.storage[slot], k) // remove key from storage when expired
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
	now := time.Now()
	start := rand.Int() & (lockStripeCount - 1)

	for offset := range lockStripeCount {
		slot := (start + offset) & (lockStripeCount - 1)
		if found, stop := s.tryPruneStripe(slot, now); found {
			return stop
		}
	}

	return true
}

func (s *mapStorage) tryPruneStripe(slot int, now time.Time) (found, stop bool) {
	s.stripes[slot].Lock()
	defer s.stripes[slot].Unlock()

	if len(s.storage[slot]) == 0 {
		return false, false
	}

	const sampleSize = 10

	sampled := 0
	expired := 0
	for k, v := range s.storage[slot] {
		if sampled == sampleSize {
			break
		}

		sampled++
		if v.ExpiresAt() != nil && now.After(*v.ExpiresAt()) {
			delete(s.storage[slot], k)
			expired++
		}
	}

	return true, expired*4 < sampled
}

func (s *mapStorage) stripe(k string) uint64 {
	return xxh3.HashString(k) & (lockStripeCount - 1)
}
