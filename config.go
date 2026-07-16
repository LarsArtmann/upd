package upd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	ProgramName = "upd"
	ProgramDesc = "Upgrade NPM Package Dependencies"
	ProgramURL  = "https://github.com/LarsArtmann/upd"

	defaultConcurrency = 8
	defaultRetries     = 3
	defaultRegistryURL = "https://registry.npmjs.org"
	defaultTimeout     = 20 * time.Second
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

// NewCommand builds the root cobra command for upd.
// The run callback receives the signal-aware context provided by fang and the
// parsed configuration. The returned Config is the same instance passed to the
// callback, so callers can inspect it after ParseFlags in tests.
func NewCommand(runE func(context.Context, *Config) error) (*cobra.Command, *Config) {
	cfg := DefaultConfig()
	cmd := &cobra.Command{
		Use:   ProgramName,
		Short: ProgramDesc,
		Long: fmt.Sprintf(`%s while preserving original JSON formatting, key order, and whitespace.

%s`, ProgramDesc, ProgramURL),
		Example: fmt.Sprintf(`  # Upgrade all dependencies in package.json
  %s

  # Preview changes without writing
  %s -n

  # Only upgrade packages matching a pattern
  %s react*`, ProgramName, ProgramName, ProgramName),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Patterns = args

			return runE(cmd.Context(), cfg)
		},
	}

	bindFlags(cmd, cfg)
	cmd.CompletionOptions.HiddenDefaultCmd = true

	return cmd, cfg
}

func bindFlags(cmd *cobra.Command, cfg *Config) {
	flags := cmd.Flags()

	flags.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, "quiet operation (no upgrade output)")
	flags.BoolVarP(&cfg.Nop, "nop", "n", cfg.Nop, "no operation (do not modify package.json)")
	flags.BoolVar(&cfg.Nop, "dry-run", cfg.Nop, "alias for --nop")
	flags.BoolVarP(&cfg.NoColor, "no-color", "C", cfg.NoColor, "do not use any colors in output")
	flags.BoolVar(&cfg.NoColor, "noColor", cfg.NoColor, "alias for --no-color")
	flags.Lookup("noColor").Hidden = true
	flags.BoolVarP(&cfg.Greatest, "greatest", "g", cfg.Greatest, "use greatest version (instead of latest stable)")
	flags.BoolVarP(&cfg.All, "all", "a", cfg.All, "show all packages (not just updated ones)")
	flags.BoolVarP(&cfg.PinLatest, "pin-latest", "P", cfg.PinLatest, "pin \"latest\" tag to exact semver version")
	flags.BoolVar(&cfg.JSON, "json", cfg.JSON, "emit machine-readable JSON to stdout instead of the table")
	flags.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "show full error chains (useful for debugging)")
	flags.StringVarP(&cfg.File, "file", "f", cfg.File, "package configuration file (default: package.json)")
	flags.StringVarP(&cfg.Registry, "registry", "r", cfg.Registry, "NPM registry base URL")
	flags.IntVarP(&cfg.Concurrency, "concurrency", "c", cfg.Concurrency, "concurrent NPM registry connections")
	flags.IntVar(&cfg.Retries, "retries", cfg.Retries, "max retries for transient registry failures")
	flags.DurationVarP(&cfg.Timeout, "timeout", "t", cfg.Timeout, "per-request timeout (e.g. 30s)")
	flags.BoolP("version", "V", false, "version for "+ProgramName)
}

// ParseFlags parses CLI arguments into a Config without executing the command.
// It is kept for backwards compatibility and for tests. If help or version is
// requested, it returns ErrHelp or ErrVersion.
func ParseFlags(args []string) (*Config, error) {
	cmd, cfg := NewCommand(func(context.Context, *Config) error { return nil })
	cmd.SetArgs(args)

	err := cmd.ParseFlags(args)
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil, ErrHelp
		}

		return nil, errorfamily.WrapRejection(err, "cli.parse_flags", "parse flags")
	}

	if flag := cmd.Flag("version"); flag != nil && flag.Changed {
		return nil, ErrVersion
	}

	cfg.Patterns = cmd.Flags().Args()

	return cfg, nil
}
