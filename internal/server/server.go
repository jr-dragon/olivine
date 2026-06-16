package server

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"net"
	"sync"
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

		conns: make(map[net.Conn]struct{}),
	}
}

type simpleSrv struct {
	handler  Handler
	restorer Restorer

	listener   net.Listener
	inShutdown atomic.Bool

	wg    sync.WaitGroup
	mu    sync.Mutex
	conns map[net.Conn]struct{}
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

		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()
		s.wg.Go(func() { s.serve(conn) })
	}
}

func (s *simpleSrv) serve(conn net.Conn) {
	defer conn.Close()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		delete(s.conns, conn)
	}()

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

	var errs []error
	if err := s.listener.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := s.closeConns(); err != nil {
		errs = append(errs, err)
	}

	s.wg.Wait()

	return errors.Join(errs...)
}

func (s *simpleSrv) closeConns() error {
	s.mu.Lock()
	conns := maps.Clone(s.conns)
	s.mu.Unlock()

	var errs []error
	for conn := range conns {
		errs = append(errs, conn.Close())
	}

	return errors.Join(errs...)
}
