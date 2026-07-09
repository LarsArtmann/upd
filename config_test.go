package upd

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"
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

	if cfg.Registry != defaultRegistryURL {
		t.Errorf("Registry = %q, want %q", cfg.Registry, defaultRegistryURL)
	}

	if cfg.Retries != defaultRetries {
		t.Errorf("Retries = %d, want %d", cfg.Retries, defaultRetries)
	}

	if cfg.Timeout != defaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, defaultTimeout)
	}

	if cfg.JSON {
		t.Error("JSON should default to false")
	}

	if cfg.Verbose {
		t.Error("Verbose should default to false")
	}
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

func TestParseFlagsDryRunAlias(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--dry-run"})

	assertFlagTrue(t, "Nop via --dry-run", cfg.Nop)
}

func TestParseFlagsShortDryRun(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-n"})

	assertFlagTrue(t, "Nop via -n", cfg.Nop)
}

func TestParseFlagsRegistryFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--registry", "https://my-registry.example.com"})

	if cfg.Registry != "https://my-registry.example.com" {
		t.Errorf("Registry = %q, want custom URL", cfg.Registry)
	}
}

func TestParseFlagsShortRegistryFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-r", "https://npm.fork.io"})

	if cfg.Registry != "https://npm.fork.io" {
		t.Errorf("Registry = %q, want https://npm.fork.io", cfg.Registry)
	}
}

func TestParseFlagsTimeoutFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--timeout", "45s"})

	if cfg.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", cfg.Timeout)
	}
}

func TestParseFlagsShortTimeoutFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-t", "10s"})

	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", cfg.Timeout)
	}
}

func TestParseFlagsRetriesFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--retries", "5"})

	if cfg.Retries != 5 {
		t.Errorf("Retries = %d, want 5", cfg.Retries)
	}
}

func TestParseFlagsJSONFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--json"})

	assertFlagTrue(t, "JSON", cfg.JSON)
}

func TestParseFlagsVerboseFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--verbose"})

	assertFlagTrue(t, "Verbose", cfg.Verbose)
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

// withoutNoColorEnv ensures NO_COLOR is unset for the duration of the test,
// restoring the original value on cleanup.
func withoutNoColorEnv(t *testing.T) {
	t.Helper()

	// Unset NO_COLOR, then use t.Setenv with a placeholder to register
	// automatic restoration. t.Setenv saves the "unset" state and will
	// unset again on cleanup. We immediately re-unset for the test body.
	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatalf("unset NO_COLOR: %v", err)
	}

	t.Setenv("NO_COLOR", "")

	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatalf("unset NO_COLOR after Setenv: %v", err)
	}
}

func TestShouldDisableColorWithNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer

	if !ShouldDisableColor(&buf) {
		t.Error("expected true when NO_COLOR is set")
	}
}

func TestShouldDisableColorNonFileWriterWithoutNoColor(t *testing.T) {
	withoutNoColorEnv(t)

	var buf bytes.Buffer

	if ShouldDisableColor(&buf) {
		t.Error("expected false for non-file writer without NO_COLOR")
	}
}

func TestShouldDisableColorPipedFileDetectedAsNonTTY(t *testing.T) {
	withoutNoColorEnv(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	t.Cleanup(func() {
		_ = r.Close()
		_ = w.Close()
	})

	if !ShouldDisableColor(w) {
		t.Error("expected true for piped *os.File (non-character-device)")
	}
}
