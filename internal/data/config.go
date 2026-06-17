package data

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrCommand = errors.New("invalid command")
)

const (
	AOFFsyncNo = iota + 1
	AOFFsyncAlways
	AOFFsyncEverySec
)

type Config struct {
	AOFEnabled bool
	AOFFsync   int
}

func NewConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return NewConfigFromBytes(content)
}

func NewConfigFromBytes(content []byte) (*Config, error) {
	cfg := Config{}
	rd := bytes.NewBuffer(append(content, '\n'))
	ln := 0
	for {
		ln++
		line, err := rd.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return &cfg, nil
			}
			return &cfg, fmt.Errorf("%w: (line: %d): %w", ErrCommand, ln, err)
		}

		if err := parse(&cfg, line); err != nil {
			return &cfg, fmt.Errorf("%w: (line: %d): %w", ErrCommand, ln, err)
		}
	}
}

func parse(cfg *Config, line string) error {
	parsed := strings.Fields(line)
	if len(parsed) == 0 || strings.HasPrefix(parsed[0], "#") {
		return nil
	}

	if len(parsed) == 1 {
		return errors.New("missing config command argument")
	}
	cmd, arg := parsed[0], parsed[1]

	switch cmd {
	case "appendonly":
		if arg == "yes" {
			cfg.AOFEnabled = true
		}
	case "appendfsync":
		switch arg {
		case "no":
			cfg.AOFFsync = AOFFsyncNo
		case "always":
			cfg.AOFFsync = AOFFsyncAlways
		case "everysec":
			cfg.AOFFsync = AOFFsyncEverySec
		}
	}

	return nil
}
