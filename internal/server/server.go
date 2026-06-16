package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"sync/atomic"

	"olivine/pkg/resp"
)

var (
	ErrServerClosed = errors.New("resp: server closed")
)

type Server interface {
	ListenAndServe() error
	RestoreFromDisk() error

	Close() error
}

func NewServer(handler Handler, restorer Restorer) Server {
	return &simpleSrv{
		handler:  handler,
		restorer: restorer,
	}
}

type simpleSrv struct {
	handler  Handler
	restorer Restorer

	listener   net.Listener
	inShutdown atomic.Bool
}

func (s *simpleSrv) RestoreFromDisk() error {
	return s.restorer.LoadFromDisk()
}

func (s *simpleSrv) ListenAndServe() (err error) {
	if s.listener, err = net.Listen("tcp", ":16879"); err != nil {
		return
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.inShutdown.Load() {
				return ErrServerClosed
			}

			slog.Error("failed to accept connection:", slog.Any("error", err))
			return err
		}

		go s.serve(conn)
	}
}

func (s *simpleSrv) serve(conn net.Conn) {
	defer conn.Close()

	rd := resp.NewReader(conn)

	for {
		ret, err := s.handler.ServeRESP(context.Background(), rd)
		if err != nil && errors.Is(err, ErrServer) {
			return
		}
		if ret == nil {
			ret = resp.NewNullBulkString()
		}

		if _, err := conn.Write(ret.Marshal()); err != nil {
			slog.Error("failed to write to conn:", slog.Any("error", err))
			return
		}
	}
}

func (s *simpleSrv) Close() error {
	s.inShutdown.Store(true)
	if s.listener == nil {
		return nil
	}

	return s.listener.Close()
}
