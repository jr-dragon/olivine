package server

import (
	"context"
	"errors"
	"log/slog"
	"net"

	"olivine/pkg/resp"
)

type Server interface {
	ListenAndServe() error
}

func NewServer(handler Handler) Server {
	return &simpleSrv{
		handler: handler,
	}
}

type simpleSrv struct {
	handler Handler

	listener net.Listener
}

func (s *simpleSrv) ListenAndServe() (err error) {
	if s.listener, err = net.Listen("tcp", ":16879"); err != nil {
		return
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection:", slog.Any("error", err))
			return err
		}

		go s.serve(conn)
	}
}

func (s *simpleSrv) serve(conn net.Conn) {
	defer conn.Close()

	rd := resp.NewReader(conn)

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
