package data

import (
	"bufio"
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

	cfg := Config{}

	rd := bufio.NewReader(f)
	ln := 0
	for {
		ln++
		line, err := rd.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) == 0 {
				return &cfg, nil
			}
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
		}

		if err := parse(&cfg, ln, line); err != nil {
			return nil, err
		}

		if errors.Is(err, io.EOF) {
			return &cfg, nil
		}
	}
}

func parse(cfg *Config, ln int, line string) error {
	parsed := strings.Fields(line)
	if len(parsed) == 0 || strings.HasPrefix(parsed[0], "#") {
		return nil
	}

	if len(parsed) != 2 {
		return fmt.Errorf("%w: (line: %d): %+v", ErrCommand, ln, parsed)
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
