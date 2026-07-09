package upd

import (
	"errors"
	"testing"
)

func assertConcurrency(t *testing.T, cfg *Config, want int) {
	t.Helper()

	if cfg.Concurrency != want {
		t.Errorf("Concurrency = %d, want %d", cfg.Concurrency, want)
	}
}

func assertFile(t *testing.T, cfg *Config, want string) {
	t.Helper()

	if cfg.File != want {
		t.Errorf("File = %q, want %s", cfg.File, want)
	}
}

func assertFlagTrue(t *testing.T, name string, got bool) {
	t.Helper()

	if !got {
		t.Errorf("%s should be true", name)
	}
}

func assertFlagFalse(t *testing.T, name string, got bool) {
	t.Helper()

	if got {
		t.Errorf("%s should default to false", name)
	}
}

func assertErr(t *testing.T, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("expected %v, got %v", target, err)
	}
}

func mustParseFlags(t *testing.T, args []string) *Config {
	t.Helper()

	cfg, err := ParseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return cfg
}

func TestParseFlagsDefaults(t *testing.T) {
	cfg := mustParseFlags(t, []string{})

	assertFile(t, cfg, "package.json")
	assertConcurrency(t, cfg, 8)
	assertFlagFalse(t, "Greatest", cfg.Greatest)
	assertFlagFalse(t, "All", cfg.All)
	assertFlagFalse(t, "Quiet", cfg.Quiet)
	assertFlagFalse(t, "Nop", cfg.Nop)
	assertFlagFalse(t, "NoColor", cfg.NoColor)
	assertFlagFalse(t, "PinLatest", cfg.PinLatest)
}

func TestParseFlagsShortFlags(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-n", "-C", "-g", "-a", "-q", "-P", "-c", "16", "react*"})

	assertFlagTrue(t, "Nop", cfg.Nop)
	assertFlagTrue(t, "NoColor", cfg.NoColor)
	assertFlagTrue(t, "Greatest", cfg.Greatest)
	assertFlagTrue(t, "All", cfg.All)
	assertFlagTrue(t, "Quiet", cfg.Quiet)
	assertFlagTrue(t, "PinLatest", cfg.PinLatest)
	assertConcurrency(t, cfg, 16)

	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "react*" {
		t.Errorf("Patterns = %v, want [react*]", cfg.Patterns)
	}
}

func TestParseFlagsLongFlags(t *testing.T) {
	cfg := mustParseFlags(t, []string{
		"--nop", "--noColor", "--greatest", "--all", "--pin-latest", "--concurrency", "4", "--file", "other.json",
	})

	assertFlagTrue(t, "Nop", cfg.Nop)
	assertFlagTrue(t, "NoColor", cfg.NoColor)
	assertFlagTrue(t, "Greatest", cfg.Greatest)
	assertFlagTrue(t, "All", cfg.All)
	assertFlagTrue(t, "PinLatest", cfg.PinLatest)
	assertConcurrency(t, cfg, 4)
	assertFile(t, cfg, "other.json")
}

func TestParseFlagsMultiplePatterns(t *testing.T) {
	cfg := mustParseFlags(t, []string{"react*", "!react-dom", "lodash"})

	if len(cfg.Patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(cfg.Patterns))
	}

	expected := []string{"react*", "!react-dom", "lodash"}
	for i, p := range expected {
		if cfg.Patterns[i] != p {
			t.Errorf("Patterns[%d] = %q, want %q", i, cfg.Patterns[i], p)
		}
	}
}

func TestParseFlagsHelpAndVersion(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		target error
	}{
		{"short help", []string{"-h"}, ErrHelp},
		{"long help", []string{"--help"}, ErrHelp},
		{"version", []string{"-V"}, ErrVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFlags(tt.args)
			assertErr(t, err, tt.target)
		})
	}
}

func TestUserAgent(t *testing.T) {
	cfg := DefaultConfig()

	ua := cfg.UserAgent()
	if ua != "upd/dev" {
		t.Errorf("UserAgent = %q, want upd/dev", ua)
	}
}
