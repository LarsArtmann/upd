package upd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	ProgramName = "upd"
	ProgramDesc = "Upgrade NPM Package Dependencies"
	ProgramURL  = "https://github.com/LarsArtmann/upd"

	defaultConcurrency  = 8
	defaultRetries      = 3
	defaultRegistryURL  = "https://registry.npmjs.org"
	defaultTimeout      = 20 * time.Second
	versionSeparatorLen = 40
)

// ProgramVersion is injected at build time via -ldflags="-X github.com/LarsArtmann/upd.ProgramVersion=1.2.3".
//
//nolint:gochecknoglobals
var ProgramVersion = "dev"

var (
	ErrHelp    = errors.New("help requested")
	ErrVersion = errors.New("version requested")
)

type Config struct {
	File        string
	Registry    string
	Greatest    bool
	All         bool
	Quiet       bool
	Nop         bool
	NoColor     bool
	PinLatest   bool
	JSON        bool
	Verbose     bool
	Concurrency int
	Retries     int
	Timeout     time.Duration
	Patterns    []string
}

func DefaultConfig() *Config {
	return &Config{
		File:        "package.json",
		Registry:    defaultRegistryURL,
		Greatest:    false,
		All:         false,
		Quiet:       false,
		Nop:         false,
		NoColor:     false,
		PinLatest:   false,
		JSON:        false,
		Verbose:     false,
		Concurrency: defaultConcurrency,
		Retries:     defaultRetries,
		Timeout:     defaultTimeout,
		Patterns:    nil,
	}
}

func (c *Config) UserAgent() string {
	return ProgramName + "/" + ProgramVersion
}

// ShouldDisableColor returns true if ANSI color codes should be suppressed.
// It honors the NO_COLOR environment variable (https://no-color.org/) and
// detects non-TTY writers (e.g. piped or redirected output).
func ShouldDisableColor(w io.Writer) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}

	if f, ok := w.(*os.File); ok {
		info, err := f.Stat()
		if err != nil {
			return true
		}

		if info.Mode()&os.ModeCharDevice == 0 {
			return true
		}
	}

	return false
}

func ParseFlags(args []string) (*Config, error) {
	cfg := DefaultConfig()
	flagSet := flag.NewFlagSet(ProgramName, flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	flagSet.Usage = func() { PrintUsage(flagSet.Output()) }

	var help, version bool
	defineBoolFlag(flagSet, &help, "h", "help", "show usage help")
	defineBoolFlag(flagSet, &version, "V", "version", "show program version information")
	defineBoolFlag(flagSet, &cfg.Quiet, "q", "quiet", "quiet operation (do not output upgrade information)")
	defineBoolFlag(flagSet, &cfg.Nop, "n", "nop", "no operation (do not modify package configuration file)")
	defineBoolFlag(flagSet, &cfg.NoColor, "C", "noColor", "do not use any colors in output")
	defineBoolFlag(flagSet, &cfg.Greatest, "g", "greatest", "use greatest version (instead of latest stable one)")
	defineBoolFlag(flagSet, &cfg.All, "a", "all", "show all packages (instead of just updated ones)")
	defineBoolFlag(flagSet, &cfg.PinLatest, "P", "pin-latest", "pin \"latest\" tag to exact semver version")
	defineBoolFlag(flagSet, &cfg.JSON, "", "json", "emit machine-readable JSON to stdout instead of the table")
	defineBoolFlag(flagSet, &cfg.Verbose, "", "verbose", "show full error chains (useful for debugging)")
	defineStringFlag(flagSet, &cfg.File, "f", "file", "package.json", "package configuration to use")
	defineStringFlag(flagSet, &cfg.Registry, "r", "registry", defaultRegistryURL, "NPM registry base URL")
	defineIntFlag(
		flagSet,
		&cfg.Concurrency,
		"c",
		"concurrency",
		defaultConcurrency,
		"number of concurrent network connections to NPM registry",
	)
	defineIntFlag(
		flagSet,
		&cfg.Retries,
		"",
		"retries",
		defaultRetries,
		"max retries for transient registry failures (429/5xx)",
	)
	defineDurationFlag(flagSet, &cfg.Timeout, "t", "timeout", defaultTimeout, "per-request timeout (e.g. 30s)")

	flagSet.BoolVar(&cfg.Nop, "dry-run", false, "alias for --nop (do not modify package.json)")

	parseErr := flagSet.Parse(args)
	if parseErr != nil {
		return nil, fmt.Errorf("parse flags: %w", parseErr)
	}

	if help {
		PrintUsage(os.Stdout)

		return nil, ErrHelp
	}

	if version {
		PrintVersion(os.Stdout)

		return nil, ErrVersion
	}

	cfg.Patterns = flagSet.Args()

	return cfg, nil
}

// defineBoolFlag registers a flag under both its short and long form so a single
// declaration covers both spellings (mirrors flag.FlagSet.BoolVar's per-name semantics).
// When short is empty, only the long form is registered.
func defineBoolFlag(flagSet *flag.FlagSet, p *bool, short, long, usage string) {
	flagSet.BoolVar(p, long, false, usage)

	if short != "" {
		flagSet.BoolVar(p, short, false, usage)
	}
}

// defineStringFlag registers a string flag under both its short and long form.
// When short is empty, only the long form is registered.
func defineStringFlag(flagSet *flag.FlagSet, p *string, short, long, def, usage string) {
	flagSet.StringVar(p, long, def, usage)

	if short != "" {
		flagSet.StringVar(p, short, def, usage)
	}
}

// defineIntFlag registers an int flag under both its short and long form.
// When short is empty, only the long form is registered.
func defineIntFlag(flagSet *flag.FlagSet, p *int, short, long string, def int, usage string) {
	flagSet.IntVar(p, long, def, usage)

	if short != "" {
		flagSet.IntVar(p, short, def, usage)
	}
}

// defineDurationFlag registers a duration flag under both its short and long form.
// When short is empty, only the long form is registered.
func defineDurationFlag(flagSet *flag.FlagSet, p *time.Duration, short, long string, def time.Duration, usage string) {
	flagSet.DurationVar(p, long, def, usage)

	if short != "" {
		flagSet.DurationVar(p, short, def, usage)
	}
}

func usageBlankLine(w io.Writer) {
	_, _ = fmt.Fprintln(w)
}

func PrintUsage(w io.Writer) {
	_, _ = fmt.Fprintf(
		w,
		"Usage: %s [-h] [-V] [-q] [-n|--dry-run] [-C] [-f <file>] [-r <registry>] [-g] [-a] [-c <concurrency>] [-P] [-t <timeout>] [--retries <n>] [--json] [--verbose] [<pattern> ...]\n",
		ProgramName,
	)
	usageBlankLine(w)
	_, _ = fmt.Fprintln(w, "Upgrade NPM package dependencies in package.json while preserving formatting.")
	usageBlankLine(w)
	_, _ = fmt.Fprintln(w, "Options:")

	lines := []struct{ short, long, desc string }{
		{"-h", "--help", "show usage help"},
		{"-V", "--version", "show program version information"},
		{"-q", "--quiet", "quiet operation (no upgrade output)"},
		{"-n", "--nop", "no operation (do not modify package.json)"},
		{"", "--dry-run", "alias for --nop"},
		{"-C", "--noColor", "do not use any colors in output"},
		{"-f", "--file", "package configuration file (default: package.json)"},
		{"-r", "--registry", "NPM registry base URL (default: registry.npmjs.org)"},
		{"-g", "--greatest", "use greatest version (instead of latest stable)"},
		{"-a", "--all", "show all packages (not just updated ones)"},
		{"-P", "--pin-latest", "pin \"latest\" tag to exact version"},
		{"-c", "--concurrency", "concurrent NPM registry connections (default: 8)"},
		{"", "--retries", "max retries for transient failures (default: 3)"},
		{"-t", "--timeout", "per-request timeout (default: 20s)"},
		{"", "--json", "machine-readable JSON output (for CI/scripts)"},
		{"", "--verbose", "show full error chains"},
	}
	for _, l := range lines {
		_, _ = fmt.Fprintf(w, "  %-4s %-16s  %s\n", l.short, l.long, l.desc)
	}

	usageBlankLine(w)
	_, _ = fmt.Fprintln(w, "Patterns:")
	_, _ = fmt.Fprintln(w, "  Positive or negative (prefixed with !) glob patterns for")
	_, _ = fmt.Fprintln(w, "  matching dependency names to update.")

	usageBlankLine(w)
	_, _ = fmt.Fprintln(w, "Exit codes:")
	_, _ = fmt.Fprintln(w, "  0   success — all dependencies resolved without errors")
	_, _ = fmt.Fprintln(w, "  1   failure — package not found, partial errors, IO error")
	_, _ = fmt.Fprintln(w, "  75  registry unavailable — transient 5xx/timeout (retryable)")
}

func PrintVersion(w io.Writer) {
	_, _ = fmt.Fprintf(w, "%s %s <%s>\n", ProgramName, ProgramVersion, ProgramURL)
	_, _ = fmt.Fprintf(w, "%s\n", ProgramDesc)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", versionSeparatorLen))
	_, _ = fmt.Fprintln(w, "Original: Copyright (c) 2015-2026 Dr. Ralf S. Engelschall")
	_, _ = fmt.Fprintln(w, "Go port:  Copyright (c) 2026 Lars Artmann — MIT License")
}
