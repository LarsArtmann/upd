package upd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	ProgramName = "upd"
	ProgramDesc = "Upgrade NPM Package Dependencies"
	ProgramURL  = "https://github.com/LarsArtmann/upd"
)

var ProgramVersion = "dev"

var (
	ErrHelp    = errors.New("help requested")
	ErrVersion = errors.New("version requested")
)

type Config struct {
	File        string
	Greatest    bool
	All         bool
	Quiet       bool
	Nop         bool
	NoColor     bool
	Concurrency int
	Patterns    []string
}

func DefaultConfig() *Config {
	return &Config{
		File:        "package.json",
		Concurrency: 8,
	}
}

func (c *Config) UserAgent() string {
	return ProgramName + "/" + ProgramVersion
}

func ParseFlags(args []string) (*Config, error) {
	cfg := DefaultConfig()
	fs := flag.NewFlagSet(ProgramName, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() { PrintUsage(fs.Output()) }

	var help, version bool
	defineBoolFlag(fs, &help, "h", "help", false, "show usage help")
	defineBoolFlag(fs, &version, "V", "version", false, "show program version information")
	defineBoolFlag(fs, &cfg.Quiet, "q", "quiet", false, "quiet operation (do not output upgrade information)")
	defineBoolFlag(fs, &cfg.Nop, "n", "nop", false, "no operation (do not modify package configuration file)")
	defineBoolFlag(fs, &cfg.NoColor, "C", "noColor", false, "do not use any colors in output")
	defineBoolFlag(fs, &cfg.Greatest, "g", "greatest", false, "use greatest version (instead of latest stable one)")
	defineBoolFlag(fs, &cfg.All, "a", "all", false, "show all packages (instead of just updated ones)")
	defineStringFlag(fs, &cfg.File, "f", "file", "package.json", "package configuration to use")
	defineIntFlag(
		fs,
		&cfg.Concurrency,
		"c",
		"concurrency",
		8,
		"number of concurrent network connections to NPM registry",
	)

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	if help {
		PrintUsage(os.Stdout)

		return nil, ErrHelp
	}

	if version {
		PrintVersion(os.Stdout)

		return nil, ErrVersion
	}

	cfg.Patterns = fs.Args()

	return cfg, nil
}

// defineBoolFlag registers a flag under both its short and long form so a single
// declaration covers both spellings (mirrors flag.FlagSet.BoolVar's per-name semantics).
func defineBoolFlag(fs *flag.FlagSet, p *bool, short, long string, def bool, usage string) {
	fs.BoolVar(p, short, def, usage)
	fs.BoolVar(p, long, def, usage)
}

// defineStringFlag registers a string flag under both its short and long form.
func defineStringFlag(fs *flag.FlagSet, p *string, short, long, def, usage string) {
	fs.StringVar(p, short, def, usage)
	fs.StringVar(p, long, def, usage)
}

// defineIntFlag registers an int flag under both its short and long form.
func defineIntFlag(fs *flag.FlagSet, p *int, short, long string, def int, usage string) {
	fs.IntVar(p, short, def, usage)
	fs.IntVar(p, long, def, usage)
}

func PrintUsage(w io.Writer) {
	fmt.Fprintf(
		w,
		"Usage: %s [-h] [-V] [-q] [-n] [-C] [-f <file>] [-g] [-a] [-c <concurrency>] [<pattern> ...]\n",
		ProgramName,
	)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Upgrade NPM package dependencies in package.json while preserving formatting.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Options:")

	lines := []struct{ short, long, desc string }{
		{"-h", "--help", "show usage help"},
		{"-V", "--version", "show program version information"},
		{"-q", "--quiet", "quiet operation (no upgrade output)"},
		{"-n", "--nop", "no operation (do not modify package.json)"},
		{"-C", "--noColor", "do not use any colors in output"},
		{"-f", "--file", "package configuration file (default: package.json)"},
		{"-g", "--greatest", "use greatest version (instead of latest stable)"},
		{"-a", "--all", "show all packages (not just updated ones)"},
		{"-c", "--concurrency", "concurrent NPM registry connections (default: 8)"},
	}
	for _, l := range lines {
		fmt.Fprintf(w, "  %-4s %-16s  %s\n", l.short, l.long, l.desc)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Patterns:")
	fmt.Fprintln(w, "  Positive or negative (prefixed with !) glob patterns for")
	fmt.Fprintln(w, "  matching dependency names to update.")
}

func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "%s %s <%s>\n", ProgramName, ProgramVersion, ProgramURL)
	fmt.Fprintf(w, "%s\n", ProgramDesc)
	fmt.Fprintln(w, strings.Repeat("-", 40))
	fmt.Fprintln(w, "Original: Copyright (c) 2015-2025 Dr. Ralf S. Engelschall")
	fmt.Fprintln(w, "Go port:  MIT License")
}
