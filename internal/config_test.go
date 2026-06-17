package internal

import (
	"errors"
	"testing"
)

func TestParseFlagsDefaults(t *testing.T) {
	cfg, err := ParseFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.File != "package.json" {
		t.Errorf("File = %q, want package.json", cfg.File)
	}

	if cfg.Concurrency != 8 {
		t.Errorf("Concurrency = %d, want 8", cfg.Concurrency)
	}

	if cfg.Greatest {
		t.Error("Greatest should default to false")
	}

	if cfg.All {
		t.Error("All should default to false")
	}

	if cfg.Quiet {
		t.Error("Quiet should default to false")
	}

	if cfg.Nop {
		t.Error("Nop should default to false")
	}

	if cfg.NoColor {
		t.Error("NoColor should default to false")
	}
}

func TestParseFlagsShortFlags(t *testing.T) {
	cfg, err := ParseFlags([]string{"-n", "-C", "-g", "-a", "-q", "-c", "16", "react*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Nop {
		t.Error("Nop should be true")
	}

	if !cfg.NoColor {
		t.Error("NoColor should be true")
	}

	if !cfg.Greatest {
		t.Error("Greatest should be true")
	}

	if !cfg.All {
		t.Error("All should be true")
	}

	if !cfg.Quiet {
		t.Error("Quiet should be true")
	}

	if cfg.Concurrency != 16 {
		t.Errorf("Concurrency = %d, want 16", cfg.Concurrency)
	}

	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "react*" {
		t.Errorf("Patterns = %v, want [react*]", cfg.Patterns)
	}
}

func TestParseFlagsLongFlags(t *testing.T) {
	cfg, err := ParseFlags(
		[]string{"--nop", "--noColor", "--greatest", "--all", "--concurrency", "4", "--file", "other.json"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Nop {
		t.Error("Nop should be true")
	}

	if !cfg.NoColor {
		t.Error("NoColor should be true")
	}

	if !cfg.Greatest {
		t.Error("Greatest should be true")
	}

	if !cfg.All {
		t.Error("All should be true")
	}

	if cfg.Concurrency != 4 {
		t.Errorf("Concurrency = %d, want 4", cfg.Concurrency)
	}

	if cfg.File != "other.json" {
		t.Errorf("File = %q, want other.json", cfg.File)
	}
}

func TestParseFlagsMultiplePatterns(t *testing.T) {
	cfg, err := ParseFlags([]string{"react*", "!react-dom", "lodash"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestParseFlagsHelp(t *testing.T) {
	_, err := ParseFlags([]string{"-h"})
	if !errors.Is(err, ErrHelp) {
		t.Errorf("expected ErrHelp, got %v", err)
	}
}

func TestParseFlagsVersion(t *testing.T) {
	_, err := ParseFlags([]string{"-V"})
	if !errors.Is(err, ErrVersion) {
		t.Errorf("expected ErrVersion, got %v", err)
	}
}

func TestParseFlagsHelpLong(t *testing.T) {
	_, err := ParseFlags([]string{"--help"})
	if !errors.Is(err, ErrHelp) {
		t.Errorf("expected ErrHelp, got %v", err)
	}
}

func TestUserAgent(t *testing.T) {
	cfg := DefaultConfig()

	ua := cfg.UserAgent()
	if ua != "upd/1.0.0" {
		t.Errorf("UserAgent = %q, want upd/1.0.0", ua)
	}
}
