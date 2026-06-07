package server

import (
	"bufio"
	"log/slog"
	"net"
	"olivine/pkg/resp"
)

type Server interface {
	ListenAndServe() error
}

func NewServer() Server {
	return &simpleSrv{}
}

type simpleSrv struct {
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

	rd := bufio.NewReader(conn)
	_, err := rd.ReadString('\n')
	if err != nil {
		slog.Error("failed to read from conn:", slog.Any("error", err))
		return
	}

	if _, err := conn.Write(resp.SimpleString("OK").Marshal()); err != nil {
		slog.Error("failed to write to conn:", slog.Any("error", err))
		return
	}
}
