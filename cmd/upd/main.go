package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"syscall"

	"charm.land/fang/v2"
	"github.com/LarsArtmann/upd"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	err := runE(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(errorfamily.ExitCode(err))
	}
}

func runE(args []string, stdout, stderr io.Writer) error {
	args, deprecated := deprecatedNoColorArgs(args)
	if deprecated {
		printWarnings(stderr, []string{noColorDeprecation})
	}

	cmd, cfg := upd.NewCommand(func(ctx context.Context, cfg *upd.Config) error {
		return executeRun(ctx, cfg, stdout, stderr)
	})
	cmd.Version = upd.ProgramVersion
	cmd.SetVersionTemplate(versionTemplate)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithoutVersion(),
		fang.WithNotifySignal(syscall.SIGINT, syscall.SIGTERM),
		fang.WithColorSchemeFunc(colorSchemeFunc(cfg)),
	)
	if err != nil {
		return errorfamily.Wrap(err, errorfamily.Classify(err), "cli.execute", "execute command")
	}

	return nil
}

const versionTemplate = `{{.Name}} {{.Version}} <https://github.com/LarsArtmann/upd>
Upgrade NPM Package Dependencies
----------------------------------------
Original: Copyright (c) 2015-2026 Dr. Ralf S. Engelschall
Go port:  Copyright (c) 2026 Lars Artmann — MIT License`

func executeRun(ctx context.Context, cfg *upd.Config, stdout, stderr io.Writer) error {
	if !cfg.NoColor {
		cfg.NoColor = upd.ShouldDisableColor(stdout)
	}

	pkg, err := upd.ReadPackageFile(cfg.File)
	if err != nil {
		return err
	}

	embedded, err := pkg.GetUpdArgs()
	if err != nil {
		return err
	}

	if len(embedded) > 0 {
		cfg.Patterns = append(embedded, cfg.Patterns...)
	}

	manifest, buildWarnings := upd.BuildManifest(pkg, cfg.Patterns, cfg.PinLatest)

	if !cfg.Quiet {
		printWarnings(stderr, append(cfg.EnvWarnings(), buildWarnings...))
	}

	toCheck := manifest.ToCheck()
	engine := upd.NewEngine(cfg)

	showProgress := !cfg.Quiet && len(toCheck) > 0

	reporter := upd.NewProgressReporter(stderr, len(toCheck), cfg.NoColor)
	if showProgress {
		reporter.Start()
		engine = engine.WithReporter(reporter)
	}

	results := engine.FetchAll(ctx, toCheck)

	if showProgress {
		reporter.Finish()
	}

	updates, errCount := engine.ApplyUpdates(manifest, results, pkg)

	return finalizeRun(cfg, manifest, pkg, updates, errCount, stdout)
}

func finalizeRun(
	cfg *upd.Config,
	manifest upd.Manifest,
	pkg *upd.PackageFile,
	updates, errCount int,
	stdout io.Writer,
) error {
	if !cfg.Quiet {
		if cfg.JSON {
			err := upd.RenderJSON(stdout, manifest)
			if err != nil {
				return err
			}
		} else {
			renderer := upd.NewRenderer(stdout, upd.RendererOptions{NoColor: cfg.NoColor, Verbose: cfg.Verbose})
			renderer.RenderTable(manifest, updates, errCount, cfg.All)
		}
	}

	if updates > 0 && !cfg.Nop {
		err := pkg.Write(cfg.File)
		if err != nil {
			return err
		}
	}

	if errCount > 0 {
		return upd.ErrPartialFailure.WithContextf("error_count", "%d", errCount)
	}

	return nil
}

const warningLine = "\x1b[33mWARNING:\x1b[0m %s\n"

// noColorDeprecation is printed when the legacy --noColor alias is used.
// The alias is accepted until v2 but no longer appears in help, man pages,
// or shell completions.
const noColorDeprecation = "--noColor is deprecated and will be removed in v2.0.0; use --no-color instead."

// deprecatedNoColorArgs maps the legacy --noColor alias onto --no-color so
// the alias keeps working without being a registered flag (which leaked it
// into man pages and completions). It reports whether any argument was
// rewritten.
func deprecatedNoColorArgs(args []string) ([]string, bool) {
	rewritten := make([]string, 0, len(args))

	deprecated := false

	for _, arg := range args {
		switch arg {
		case "--noColor":
			rewritten = append(rewritten, "--no-color")
			deprecated = true
		case "--noColor=true":
			rewritten = append(rewritten, "--no-color=true")
			deprecated = true
		case "--noColor=false":
			rewritten = append(rewritten, "--no-color=false")
			deprecated = true
		default:
			rewritten = append(rewritten, arg)
		}
	}

	return rewritten, deprecated
}

// printWarnings writes warnings to stderr. Write errors are not actionable
// (a closed terminal cannot be reported to), matching the renderer's
// deliberate ignore policy.
func printWarnings(w io.Writer, warnings []string) {
	for _, msg := range warnings {
		_, _ = fmt.Fprintf(w, warningLine, msg) //nolint:erraudit
	}
}
