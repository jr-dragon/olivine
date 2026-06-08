package server

import (
	"context"
	"errors"
	"io"
	
	"olivine/internal/service"
)

type Restorer interface {
	LoadFromDisk() error
}

func NewRestorer(aof service.AOF, handler Handler) Restorer {
	return &aofRestorer{aof: aof, handler: handler}
}

type aofRestorer struct {
	aof     service.AOF
	handler Handler
}

func (r *aofRestorer) LoadFromDisk() error {
	for {
		cmd, err := r.aof.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		if _, err := r.handler.Exec(context.Background(), cmd); err != nil {
			return err
		}
	}
}
