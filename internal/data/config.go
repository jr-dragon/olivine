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
	cfg := Config{}

	rd := bufio.NewReader(f)
	ln := 0
	for {
		ln++
		line, err := rd.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return &cfg, nil
			}

			return nil, err
		}

		line = strings.Trim(line, " \n")
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parsed := strings.SplitN(line, " ", 2)
		if len(parsed) != 2 {
			return nil, fmt.Errorf("%w: (line: %d): %+v", ErrCommand, ln, parsed)
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
	}
}
