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

	Shutdown(context.Context) error
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
		if err != nil {
			if s.inShutdown.Load() {
				return
			}
			if errors.Is(err, ErrServer) {
				slog.Error("failed to serve resp", slog.Any("error", err))
				return
			}
		}
		if ret == nil {
			ret = resp.NewNullBulkString()
		}

		if _, err := conn.Write(ret.Marshal()); err != nil {
			if s.inShutdown.Load() {
				return
			}

			slog.Error("failed to write to conn:", slog.Any("error", err))
			return
		}
	}
}

func (s *simpleSrv) Shutdown(ctx context.Context) error {
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

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return errors.Join(errs...)
	case <-ctx.Done():
		errs = append(errs, ctx.Err())
		return errors.Join(errs...)
	}
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
