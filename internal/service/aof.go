package service

import (
	"fmt"
	"olivine/pkg/resp"
	"os"
	"sync"
)

type AOF interface {
	Read() (*resp.Command, error)
	Write(*resp.Command) error
	Close() error
}

type file struct {
	f  *os.File
	rd *resp.Reader
	mu sync.Mutex
}

func (aof *file) Read() (*resp.Command, error) {
	return resp.ReadCommand(aof.rd)
}

func (aof *file) Write(v *resp.Command) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	if !v.Dirty() {
		return nil
	}

	marshaled := v.Marshal()
	n, err := aof.f.Write(marshaled)
	if err != nil {
		return err
	}
	if n != len(marshaled) {
		return fmt.Errorf("wrote length mismatch: got %d want %d", n, len(marshaled))
	}

	return nil
}

func (aof *file) Close() error {
	if aof.f == nil {
		return nil
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.f.Close()
}
