package upd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
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

	if cfg.Format != defaultFormat {
		t.Errorf("Format = %q, want default %q", cfg.Format, defaultFormat)
	}

	if cfg.Verbose {
		t.Error("Verbose should default to false")
	}
}

func assertCoreBoolFlags(t *testing.T, cfg *Config) {
	t.Helper()

	assertFlagTrue(t, "Nop", cfg.Nop)
	assertFlagTrue(t, "NoColor", cfg.NoColor)
	assertFlagTrue(t, "Greatest", cfg.Greatest)
	assertFlagTrue(t, "All", cfg.All)
	assertFlagTrue(t, "PinLatest", cfg.PinLatest)
}

func TestParseFlagsShortFlags(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-n", "-C", "-g", "-a", "-q", "-P", "-c", "16", "react*"})

	assertCoreBoolFlags(t, cfg)
	assertFlagTrue(t, "Quiet", cfg.Quiet)
	assertConcurrency(t, cfg, 16)

	if len(cfg.Patterns) != 1 || cfg.Patterns[0] != "react*" {
		t.Errorf("Patterns = %v, want [react*]", cfg.Patterns)
	}
}

func TestParseFlagsLongFlags(t *testing.T) {
	cfg := mustParseFlags(t, []string{
		"--nop", "--no-color", "--greatest", "--all", "--pin-latest", "--concurrency", "4", "--file", "other.json",
	})

	assertCoreBoolFlags(t, cfg)
	assertConcurrency(t, cfg, 4)
	assertFile(t, cfg, "other.json")
}

func TestParseFlagsDryRunAlias(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--dry-run"})

	assertFlagTrue(t, "Nop via --dry-run", cfg.Nop)
}

// TestParseFlagsNoColorAlias documents that the legacy --noColor alias is no
// longer a registered flag: the CLI rewrites it in runE (with a deprecation
// warning) so it never reaches the flag parser.
func TestParseFlagsNoColorAliasRejected(t *testing.T) {
	if _, err := ParseFlags([]string{"--noColor"}); err == nil {
		t.Error("expected --noColor to be rejected by ParseFlags (handled by CLI rewrite)")
	}
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

func TestParseFlagsFormatFlag(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--format=json"})

	if cfg.Format != FormatJSON {
		t.Errorf("Format = %q, want %q", cfg.Format, FormatJSON)
	}
}

func TestParseFlagsFormatTableExplicit(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--format=table"})

	if cfg.Format != FormatTable {
		t.Errorf("Format = %q, want %q", cfg.Format, FormatTable)
	}
}

func TestParseFlagsInvalidFormatRejected(t *testing.T) {
	_, err := ParseFlags([]string{"--format=yaml"})
	if !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("expected ErrInvalidFormat, got %v", err)
	}
}

func TestParseFlagsSilentAlias(t *testing.T) {
	cfg := mustParseFlags(t, []string{"--silent"})

	assertFlagTrue(t, "Quiet via --silent", cfg.Quiet)
}

func TestParseFlagsSilentShorthand(t *testing.T) {
	cfg := mustParseFlags(t, []string{"-s"})

	assertFlagTrue(t, "Quiet via -s", cfg.Quiet)
}

// TestJSONFlagRejectedByParser documents that the legacy --json alias is
// rewritten by the CLI layer (see cmd/upd rewriteDeprecatedArgs) instead of
// being a registered flag, so the parser no longer knows it.
func TestJSONFlagRejectedByParser(t *testing.T) {
	if _, err := ParseFlags([]string{"--json"}); err == nil {
		t.Error("expected --json to be rejected by ParseFlags (handled by CLI rewrite)")
	}
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
		{"short version", []string{"-V"}, ErrVersion},
		{"long version", []string{"--version"}, ErrVersion},
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

func TestParseFlagsEnvVars(t *testing.T) {
	tests := []struct {
		name   string
		env    map[string]string
		args   []string
		assert func(t *testing.T, cfg *Config)
	}{
		{
			name: "UPD_REGISTRY sets registry",
			env:  map[string]string{EnvRegistry: "https://env.registry.example.com"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Registry != "https://env.registry.example.com" {
					t.Errorf("Registry = %q, want env value", cfg.Registry)
				}
			},
		},
		{
			name: "UPD_FILE sets file",
			env:  map[string]string{EnvFile: "env-package.json"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertFile(t, cfg, "env-package.json")
			},
		},
		{
			name: "UPD_TIMEOUT sets timeout",
			env:  map[string]string{EnvTimeout: "45s"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Timeout != 45*time.Second {
					t.Errorf("Timeout = %v, want 45s", cfg.Timeout)
				}
			},
		},
		{
			name: "UPD_CONCURRENCY sets concurrency",
			env:  map[string]string{EnvConcurrency: "16"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertConcurrency(t, cfg, 16)
			},
		},
		{
			name: "UPD_RETRIES sets retries",
			env:  map[string]string{EnvRetries: "5"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				if cfg.Retries != 5 {
					t.Errorf("Retries = %d, want 5", cfg.Retries)
				}
			},
		},
		{
			name: "UPD_NO_COLOR sets no color",
			env:  map[string]string{EnvNoColor: "true"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertFlagTrue(t, "NoColor", cfg.NoColor)
			},
		},
		{
			name: "UPD_QUIET sets quiet",
			env:  map[string]string{EnvQuiet: "true"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertFlagTrue(t, "Quiet", cfg.Quiet)
			},
		},
		{
			name: "UPD_GREATEST sets greatest",
			env:  map[string]string{EnvGreatest: "1"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertFlagTrue(t, "Greatest", cfg.Greatest)
			},
		},
		{
			name: "CLI flag overrides env var",
			env:  map[string]string{EnvConcurrency: "4"},
			args: []string{"--concurrency", "12"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertConcurrency(t, cfg, 12)
			},
		},
		{
			name: "invalid env var ignored",
			env:  map[string]string{EnvConcurrency: "not-a-number"},
			assert: func(t *testing.T, cfg *Config) {
				t.Helper()
				assertConcurrency(t, cfg, defaultConcurrency)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg := mustParseFlags(t, tt.args)
			tt.assert(t, cfg)
		})
	}
}

func TestNewCommandMetadata(t *testing.T) {
	cmd, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cmd.Use != ProgramName {
		t.Errorf("Use = %q, want %q", cmd.Use, ProgramName)
	}

	if cmd.Short != ProgramDesc {
		t.Errorf("Short = %q, want %q", cmd.Short, ProgramDesc)
	}

	if cfg == nil {
		t.Fatal("expected non-nil Config")
	}

	if !cmd.CompletionOptions.HiddenDefaultCmd {
		t.Error("expected default completion command to be hidden")
	}

	flags := []string{
		"quiet", "silent", "nop", "dry-run", "no-color", "greatest", "all",
		"pin-latest", "format", "verbose", "file", "registry", "concurrency",
		"retries", "timeout", "version",
	}

	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag %q", name)
		}
	}
}

func TestApplyEnvFlagsInvalidValuesProduceWarnings(t *testing.T) {
	t.Setenv("UPD_TIMEOUT", "30")
	t.Setenv("UPD_CONCURRENCY", "not-a-number")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	warnings := cfg.EnvWarnings()
	if len(warnings) != 2 {
		t.Fatalf("expected 2 env warnings, got %d: %q", len(warnings), warnings)
	}

	joined := strings.Join(warnings, "\n")
	if !strings.Contains(joined, "UPD_TIMEOUT") || !strings.Contains(joined, "UPD_CONCURRENCY") {
		t.Errorf("warnings should name the offending env vars: %q", joined)
	}

	if cfg.Timeout != defaultTimeout {
		t.Errorf("Timeout = %v, want default %v (invalid env ignored)", cfg.Timeout, defaultTimeout)
	}

	if cfg.Concurrency != defaultConcurrency {
		t.Errorf("Concurrency = %d, want default %d", cfg.Concurrency, defaultConcurrency)
	}
}

func TestApplyEnvFlagsValidValuesProduceNoWarnings(t *testing.T) {
	t.Setenv("UPD_TIMEOUT", "30s")
	t.Setenv("UPD_CONCURRENCY", "2")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if warnings := cfg.EnvWarnings(); len(warnings) != 0 {
		t.Errorf("expected no env warnings for valid values, got %q", warnings)
	}

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s from UPD_TIMEOUT", cfg.Timeout)
	}

	if cfg.Concurrency != 2 {
		t.Errorf("Concurrency = %d, want 2 from UPD_CONCURRENCY", cfg.Concurrency)
	}
}

func TestApplyEnvFormatSetsJSONOutput(t *testing.T) {
	t.Setenv("UPD_FORMAT", "json")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cfg.Format != FormatJSON {
		t.Errorf("Format = %q, want %q from UPD_FORMAT", cfg.Format, FormatJSON)
	}

	if warnings := cfg.EnvWarnings(); len(warnings) != 0 {
		t.Errorf("expected no warnings for UPD_FORMAT, got %q", warnings)
	}
}

func TestApplyDeprecatedEnvJSONWarnsAndStillWorks(t *testing.T) {
	t.Setenv("UPD_JSON", "true")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cfg.Format != FormatJSON {
		t.Errorf("Format = %q, want %q from deprecated UPD_JSON", cfg.Format, FormatJSON)
	}

	warnings := cfg.EnvWarnings()
	if len(warnings) != 1 {
		t.Fatalf("expected 1 deprecation warning, got %d: %q", len(warnings), warnings)
	}

	if !strings.Contains(warnings[0], "UPD_JSON is deprecated") || !strings.Contains(warnings[0], "UPD_FORMAT=json") {
		t.Errorf("warning should point at UPD_FORMAT=json, got: %q", warnings[0])
	}
}

func TestApplyDeprecatedEnvJSONFalseIsIgnored(t *testing.T) {
	t.Setenv("UPD_JSON", "false")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cfg.Format != defaultFormat {
		t.Errorf("Format = %q, want default", cfg.Format)
	}

	if warnings := cfg.EnvWarnings(); len(warnings) != 0 {
		t.Errorf("expected no warnings for UPD_JSON=false, got %q", warnings)
	}
}

func TestApplyDeprecatedEnvJSONInvalidValueWarns(t *testing.T) {
	t.Setenv("UPD_JSON", "not-a-bool")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cfg.Format != defaultFormat {
		t.Errorf("Format = %q, want default (invalid env ignored)", cfg.Format)
	}

	warnings := cfg.EnvWarnings()
	if len(warnings) != 1 || !strings.Contains(warnings[0], "UPD_JSON") {
		t.Errorf("expected 1 warning naming UPD_JSON, got %q", warnings)
	}
}

func TestApplyEnvFormatWinsOverDeprecatedJSON(t *testing.T) {
	t.Setenv("UPD_JSON", "true")
	t.Setenv("UPD_FORMAT", "table")

	_, cfg := NewCommand(func(context.Context, *Config) error { return nil })

	if cfg.Format != FormatTable {
		t.Errorf("Format = %q, want %q (UPD_FORMAT takes precedence)", cfg.Format, FormatTable)
	}
}

func TestSuggestFlagTypo(t *testing.T) {
	t.Parallel()

	candidates := []string{
		"quiet", "silent", "nop", "dry-run", "no-color", "greatest", "all", "pin-latest",
		"format", "verbose", "file", "registry", "concurrency", "retries", "timeout",
	}

	tests := []struct {
		input string
		want  string
		found bool
	}{
		{"forma", "format", true},
		{"format", "format", true},
		{"verbos", "verbose", true},
		{"qiet", "quiet", true},
		{"silen", "silent", true},
		{"no-colour", "no-color", true},
		{"pin-late", "pin-latest", true},
		{"registryy", "registry", true},
		{"zzzzzzz", "", false},
	}

	for _, tc := range tests {
		got, found := suggestFlag(tc.input, candidates)
		if found != tc.found || got != tc.want {
			t.Errorf("suggestFlag(%q) = (%q, %v), want (%q, %v)", tc.input, got, found, tc.want, tc.found)
		}
	}
}

func TestParseFlagsUnknownFlagSuggestsAlternative(t *testing.T) {
	_, err := ParseFlags([]string{"--forma"})
	if err == nil {
		t.Fatal("expected unknown flag error")
	}

	if !strings.Contains(err.Error(), "Did you mean --format?") {
		t.Errorf("error should contain a suggestion, got: %v", err)
	}
}

func TestParseFlagsRemovedJSONFlagSuggestsReplacement(t *testing.T) {
	_, err := ParseFlags([]string{"--jso"})
	if err == nil {
		t.Fatal("expected unknown flag error")
	}

	if !strings.Contains(err.Error(), "Did you mean --format=json?") {
		t.Errorf("error should suggest the --format replacement, got: %v", err)
	}

	if !strings.Contains(err.Error(), "--json is deprecated") {
		t.Errorf("error should note the deprecation, got: %v", err)
	}
}

func TestParseFlagsUnknownFlagWithoutSuggestion(t *testing.T) {
	_, err := ParseFlags([]string{"--zzzzzzz"})
	if err == nil {
		t.Fatal("expected unknown flag error")
	}

	if strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("error should not contain a suggestion for a distant typo, got: %v", err)
	}
}
