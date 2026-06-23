package service

import (
	"fmt"
	"os"
	"sync"

	"olivine/internal/data"
	"olivine/pkg/resp"
)

type AOF interface {
	Read() (*resp.Command, error)
	Write(*resp.Command) error
	Sync() error
	Close() error
}

func NewAOF(cfg *data.Config, path string) (AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	return &file{cfg: cfg, f: f, rd: resp.NewReader(f)}, nil
}

type file struct {
	cfg *data.Config

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

	marshaled := v.MarshalAOF()
	n, err := aof.f.Write(marshaled)
	if err != nil {
		return err
	}
	if n != len(marshaled) {
		return fmt.Errorf("wrote length mismatch: got %d want %d", n, len(marshaled))
	}

	if aof.cfg.AOFFsync == data.AOFFsyncAlways {
		return aof.f.Sync()
	}

	return nil
}

func (aof *file) Sync() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.f.Sync()
}

func (aof *file) Close() error {
	if aof.f == nil {
		return nil
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.f.Close()
}
