package data

import "testing"

func TestParseAppendOnly(t *testing.T) {
	cfg := Config{}

	if err := parse(&cfg, 1, "appendonly yes"); err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	if !cfg.AOFEnabled {
		t.Fatal("cfg.AOFEnabled = false, want true")
	}
}

func TestParseAppendFsync(t *testing.T) {
	tests := []struct {
		name string
		line string
		want int
	}{
		{name: "no", line: "appendfsync no", want: AOFFsyncNo},
		{name: "always", line: "appendfsync always", want: AOFFsyncAlways},
		{name: "everysec", line: "appendfsync everysec", want: AOFFsyncEverySec},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{}
			if err := parse(&cfg, 1, tt.line); err != nil {
				t.Fatalf("parse() error = %v", err)
			}
			if cfg.AOFFsync != tt.want {
				t.Fatalf("cfg.AOFFsync = %d, want %d", cfg.AOFFsync, tt.want)
			}
		})
	}
}

func TestParseFlexibleWhitespace(t *testing.T) {
	cfg := Config{}

	if err := parse(&cfg, 1, "\tappendfsync   everysec\r\n"); err != nil {
		t.Fatalf("parse() error = %v", err)
	}

	if cfg.AOFFsync != AOFFsyncEverySec {
		t.Fatalf("cfg.AOFFsync = %d, want %d", cfg.AOFFsync, AOFFsyncEverySec)
	}
}

func TestParseSkipsEmptyLineAndComment(t *testing.T) {
	cfg := Config{}

	if err := parse(&cfg, 1, " \t\r\n"); err != nil {
		t.Fatalf("parse() empty line error = %v", err)
	}
	if err := parse(&cfg, 2, "# appendonly yes"); err != nil {
		t.Fatalf("parse() comment error = %v", err)
	}

	if cfg.AOFEnabled {
		t.Fatal("cfg.AOFEnabled = true, want false")
	}
	if cfg.AOFFsync != 0 {
		t.Fatalf("cfg.AOFFsync = %d, want 0", cfg.AOFFsync)
	}
}
