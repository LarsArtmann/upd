package upd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
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
	defaultFormat      = FormatTable
)

// Output formats accepted by the --format flag.
const (
	FormatTable = "table"
	FormatJSON  = "json"
)

const (
	EnvQuiet       = "UPD_QUIET"
	EnvNop         = "UPD_NOP"
	EnvDryRun      = "UPD_DRY_RUN"
	EnvNoColor     = "UPD_NO_COLOR"
	EnvGreatest    = "UPD_GREATEST"
	EnvAll         = "UPD_ALL"
	EnvPinLatest   = "UPD_PIN_LATEST"
	EnvJSON        = "UPD_JSON"
	EnvFormat      = "UPD_FORMAT"
	EnvVerbose     = "UPD_VERBOSE"
	EnvFile        = "UPD_FILE"
	EnvRegistry    = "UPD_REGISTRY"
	EnvConcurrency = "UPD_CONCURRENCY"
	EnvRetries     = "UPD_RETRIES"
	EnvTimeout     = "UPD_TIMEOUT"
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
	File          string
	Registry      string
	Greatest      bool
	All           bool
	Quiet         bool
	Nop           bool
	NoColor       bool
	PinLatest     bool
	Format        string
	Verbose       bool
	Concurrency   int
	Retries       int
	Timeout       time.Duration
	Patterns      []string
	RegistryToken string

	envWarnings []string
}

func DefaultConfig() *Config {
	return &Config{
		File:          "package.json",
		Registry:      defaultRegistryURL,
		Greatest:      false,
		All:           false,
		Quiet:         false,
		Nop:           false,
		NoColor:       false,
		PinLatest:     false,
		Format:        defaultFormat,
		Verbose:       false,
		Concurrency:   defaultConcurrency,
		Retries:       defaultRetries,
		Timeout:       defaultTimeout,
		Patterns:      nil,
		RegistryToken: "",
		envWarnings:   nil,
	}
}

func (c *Config) UserAgent() string {
	return ProgramName + "/" + ProgramVersion
}

// Validate clamps zero or negative values to safe defaults so the engine
// cannot deadlock or hang on misconfigured callers. It mutates the receiver
// in place and returns it for chaining.
func (c *Config) Validate() *Config {
	if c.Concurrency <= 0 {
		c.Concurrency = defaultConcurrency
	}

	if c.Retries < 0 {
		c.Retries = defaultRetries
	}

	if c.Timeout <= 0 {
		c.Timeout = defaultTimeout
	}

	if c.Registry == "" {
		c.Registry = defaultRegistryURL
	}

	if c.Format == "" {
		c.Format = defaultFormat
	}

	return c
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
		// Positional arguments are dependency-name patterns; without an
		// explicit Args, cobra's legacyArgs rejects them as unknown commands
		// once fang registers its hidden man/completion subcommands.
		Args: cobra.ArbitraryArgs,
		Long: fmt.Sprintf(`%s while preserving original JSON formatting, key order, and whitespace.

Exit codes: 0 = success, 1 = failure (package not found, partial failures, IO or JSON errors), 75 = registry unavailable (transient, safe to retry).

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
	cfg.envWarnings = append(applyDeprecatedEnvJSON(cmd), applyEnvFlags(cmd)...)
	cmd.SetFlagErrorFunc(suggestFlagOnError)
	cmd.CompletionOptions.HiddenDefaultCmd = true

	return cmd, cfg
}

func bindFlags(cmd *cobra.Command, cfg *Config) {
	flags := cmd.Flags()

	flags.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, "quiet operation (no upgrade output)")
	flags.BoolVarP(&cfg.Nop, "nop", "n", cfg.Nop, "no operation (do not modify package.json)")
	flags.BoolVar(&cfg.Nop, "dry-run", cfg.Nop, "alias for --nop")
	flags.BoolVarP(&cfg.NoColor, "no-color", "C", cfg.NoColor, "do not use any colors in output")
	flags.BoolVarP(&cfg.Greatest, "greatest", "g", cfg.Greatest, "use greatest version (instead of latest stable)")
	flags.BoolVarP(&cfg.All, "all", "a", cfg.All, "show all packages (not just updated ones)")
	flags.BoolVarP(&cfg.PinLatest, "pin-latest", "P", cfg.PinLatest, "pin \"latest\" tag to exact semver version")
	flags.StringVar(&cfg.Format, "format", cfg.Format, "output format: table or json")
	flags.BoolVarP(&cfg.Quiet, "silent", "s", cfg.Quiet, "alias for --quiet")
	flags.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "show full error chains (useful for debugging)")
	flags.StringVarP(&cfg.File, "file", "f", cfg.File, "package configuration file (default: package.json)")
	flags.StringVarP(&cfg.Registry, "registry", "r", cfg.Registry, "NPM registry base URL")
	flags.IntVarP(&cfg.Concurrency, "concurrency", "c", cfg.Concurrency, "concurrent NPM registry connections")
	flags.IntVar(&cfg.Retries, "retries", cfg.Retries, "max retries for transient registry failures")
	flags.DurationVarP(&cfg.Timeout, "timeout", "t", cfg.Timeout, "per-request timeout (e.g. 30s)")
	flags.BoolP("version", "V", false, "version for "+ProgramName)
}

// envFlag maps a flag name to the environment variable that can override it.
// The env var is applied before flag parsing, so explicit CLI flags still win.
type envFlag struct {
	flag string
	env  string
}

// applyEnvFlags applies environment variables to their bound flags before Cobra
// parses CLI arguments. If an env var is set and its value is valid for the flag
// type, it becomes the flag's default value; an explicit CLI flag then overrides
// it. Invalid env var values are ignored so the program's built-in defaults
// remain in effect; each ignored value is reported as a warning string.
func applyEnvFlags(cmd *cobra.Command) []string {
	var warnings []string

	for _, mapping := range envFlagMappings() {
		value, ok := os.LookupEnv(mapping.env)
		if !ok {
			continue
		}

		flag := cmd.Flags().Lookup(mapping.flag)
		if flag == nil {
			continue
		}

		original := flag.Value.String()
		if err := cmd.Flags().Set(mapping.flag, value); err != nil {
			warnings = append(warnings, fmt.Sprintf(
				"invalid env var %s=%q is ignored (using the default): %v",
				mapping.env, value, err,
			))

			// Restore the value read from this same flag; Set cannot fail here.
			_ = cmd.Flags().Set(mapping.flag, original) //nolint:erraudit
		}
	}

	return warnings
}

// envFlagMappings returns the list of flags that may be read from environment
// variables. The --version flag is intentionally omitted.
func envFlagMappings() []envFlag {
	return []envFlag{
		{"quiet", EnvQuiet},
		{"nop", EnvNop},
		{"dry-run", EnvDryRun},
		{"no-color", EnvNoColor},
		{"greatest", EnvGreatest},
		{"all", EnvAll},
		{"pin-latest", EnvPinLatest},
		{"format", EnvFormat},
		{"verbose", EnvVerbose},
		{"file", EnvFile},
		{"registry", EnvRegistry},
		{"concurrency", EnvConcurrency},
		{"retries", EnvRetries},
		{"timeout", EnvTimeout},
	}
}

// applyDeprecatedEnvJSON honors the deprecated UPD_JSON env var by mapping
// truthy values onto the --format flag so existing CI setups keep working.
// It returns one warning per applied or ignored value; UPD_FORMAT, applied
// afterwards by applyEnvFlags, still takes precedence.
func applyDeprecatedEnvJSON(cmd *cobra.Command) []string {
	value, ok := os.LookupEnv(EnvJSON)
	if !ok {
		return nil
	}

	jsonWanted, err := strconv.ParseBool(value)
	if err != nil {
		return []string{fmt.Sprintf(
			"invalid env var %s=%q is ignored (using the default): not a boolean value", EnvJSON, value,
		)}
	}

	if !jsonWanted {
		return nil
	}

	if err := cmd.Flags().Set("format", FormatJSON); err != nil {
		return []string{fmt.Sprintf(
			"invalid env var %s=%q is ignored (using the default): %v", EnvJSON, value, err,
		)}
	}

	return []string{fmt.Sprintf(
		"env var %s is deprecated and will be removed in v2.0.0; use %s=%s instead", EnvJSON, EnvFormat, FormatJSON,
	)}
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

		// cobra applies the flag-error hook in execute() only; apply it here
		// as well so library callers get the same suggestion.
		return nil, errorfamily.WrapRejection(suggestFlagOnError(cmd, err), "cli.parse_flags", "parse flags")
	}

	if flag := cmd.Flag("version"); flag != nil && flag.Changed {
		return nil, ErrVersion
	}

	if cfg.Format != FormatTable && cfg.Format != FormatJSON {
		return nil, ErrInvalidFormat.WithContextf("format", "%s", cfg.Format)
	}

	cfg.Patterns = cmd.Flags().Args()

	return cfg, nil
}

// EnvWarnings returns one warning per UPD_* environment variable that was set
// to an invalid value and ignored in favor of the built-in default.
func (c *Config) EnvWarnings() []string {
	return c.envWarnings
}

// suggestFlagOnError is the cobra flag-error hook: it appends a "did you mean"
// suggestion for unknown long flags so typos like --jso point at --format.
func suggestFlagOnError(cmd *cobra.Command, err error) error {
	// Every error reaching this func is a flag-parsing failure, i.e. a user
	// input problem rather than a transient fault. Classify it as Rejection so
	// a typo exits 1; the unclassified default would exit 75 and tell CI to
	// retry a command that can never succeed.
	wrapped := errorfamily.WrapRejection(err, "cli.invalid_flag", "invalid command-line usage")

	msg := err.Error()

	name, ok := strings.CutPrefix(msg, "unknown flag: --")
	if !ok || name == "" || strings.ContainsAny(name, " =") {
		return wrapped
	}

	candidates := append(flagNames(cmd), deprecatedFlagNames()...)

	suggestion, found := suggestFlag(name, candidates)
	if !found {
		return wrapped
	}

	if replacement, deprecated := deprecatedFlagSuggestions()[suggestion]; deprecated {
		return fmt.Errorf("%w\n\nDid you mean --%s? (--%s is deprecated)", wrapped, replacement, suggestion)
	}

	return fmt.Errorf("%w\n\nDid you mean --%s?", wrapped, suggestion)
}

// deprecatedFlagSuggestions maps removed flag names onto their canonical
// replacement spellings so typos of removed flags still produce a useful hint.
func deprecatedFlagSuggestions() map[string]string {
	return map[string]string{
		"json": "format=" + FormatJSON,
	}
}

func deprecatedFlagNames() []string {
	return slices.Collect(maps.Keys(deprecatedFlagSuggestions()))
}

// flagNames lists all long flag names declared on the command.
func flagNames(cmd *cobra.Command) []string {
	names := make([]string, 0, cmd.Flags().NFlag())

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		names = append(names, f.Name)
	})

	return names
}

// suggestFlag returns the closest candidate for a mistyped flag name,
// accepting edits up to Levenshtein distance 2 or a shared prefix. The second
// return value reports whether a suggestion was found.
func suggestFlag(name string, candidates []string) (string, bool) {
	best := ""
	bestDistance := maxFlagTypoDistance + 1

	for _, candidate := range candidates {
		if candidate == name {
			return candidate, true
		}

		if strings.HasPrefix(candidate, name) || strings.HasPrefix(name, candidate) {
			return candidate, true
		}

		distance := levenshteinDistance(name, candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}

	if bestDistance <= maxFlagTypoDistance {
		return best, true
	}

	return "", false
}

// maxFlagTypoDistance is the largest edit distance still considered a typo
// rather than a different word.
const maxFlagTypoDistance = 2

func levenshteinDistance(from, to string) int {
	fromRunes, toRunes := []rune(from), []rune(to)

	prev := make([]int, 0, len(toRunes)+1)
	curr := make([]int, 0, len(toRunes)+1)

	for j := range len(toRunes) + 1 {
		prev = append(prev, j)
	}

	for i := range fromRunes {
		curr = append(curr, i+1)

		for j := range toRunes {
			cost := 1
			if fromRunes[i] == toRunes[j] {
				cost = 0
			}

			curr = append(curr, min(prev[j+1]+1, curr[j]+1, prev[j]+cost))
		}

		prev, curr = curr, prev
		curr = curr[:0]
	}

	return prev[len(toRunes)]
}
