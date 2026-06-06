package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"strings"
)

type App struct{}

func (app *App) Run() error {
	slog.Info("starting olivine server...")
	listener, err := net.Listen("tcp", ":16879")
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %w", err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection:", slog.Any("error", err))
			continue
		}
		go handleConnection(connection)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		slog.Error("failed to read from conn:", slog.Any("error", err))
		return
	}

	ackmsg := strings.ToUpper(strings.ToUpper(message))
	response := "ACK: " + ackmsg
	if _, err := conn.Write([]byte(response)); err != nil {
		slog.Error("failed to write to conn:", slog.Any("error", err))
		return
	}
}
