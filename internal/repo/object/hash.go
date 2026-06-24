package object

import (
	"sync"
	"time"
)

type Hash struct {
	base

	mu  sync.RWMutex
	val map[string]string
}

func NewHash(k string, expiresAt *time.Time) *Hash {
	return &Hash{
		val:  make(map[string]string),
		base: base{key: k, expiresAt: expiresAt},
	}
}
