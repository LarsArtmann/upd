package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	args, warnings := rewriteDeprecatedArgs(args)
	printWarnings(stderr, warnings)

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
	if cfg.Format != upd.FormatTable && cfg.Format != upd.FormatJSON {
		return upd.ErrInvalidFormat.WithContextf("format", "%s", cfg.Format)
	}

	if !cfg.NoColor {
		cfg.NoColor = upd.ShouldDisableColor(stdout)
	}

	npmrc, npmrcWarnings := upd.LoadNpmrc(filepath.Dir(cfg.File))
	cfg.ApplyNpmrc(npmrc)

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
		printWarnings(stderr, append(append(cfg.EnvWarnings(), buildWarnings...), npmrcWarnings...))
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
		if cfg.Format == upd.FormatJSON {
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
		return aggregateFailures(manifest, errCount)
	}

	return nil
}

// aggregateFailures joins the overall partial-failure signal with every
// concrete package error so programmatic consumers can inspect all failures
// through the returned error instead of parsing rendered output.
func aggregateFailures(manifest upd.Manifest, errCount int) error {
	errs := make([]error, 0, errCount+1)
	errs = append(errs, upd.ErrPartialFailure.WithContextf("error_count", "%d", errCount))

	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if spec.State == upd.StateError && spec.Err != nil {
				errs = append(errs, spec.Err)
			}
		}
	}

	return errors.Join(errs...)
}

const warningLine = "\x1b[33mWARNING:\x1b[0m %s\n"

// noColorDeprecation is printed when the legacy --noColor alias is used.
// The alias is accepted until v2 but no longer appears in help, man pages,
// or shell completions.
const noColorDeprecation = "--noColor is deprecated and will be removed in v2.0.0; use --no-color instead."

// jsonDeprecation is printed when the legacy --json alias is used. The alias
// is accepted until v2 but no longer appears in help, man pages, or shell
// completions.
const jsonDeprecation = "--json is deprecated and will be removed in v2.0.0; use --format=json instead."

// rewriteDeprecatedArgs maps deprecated flag spellings onto their canonical
// forms so the aliases keep working without being registered flags (which
// leaked them into man pages and completions). It returns the rewritten
// arguments and one deprecation warning per rewritten flag.
func rewriteDeprecatedArgs(args []string) ([]string, []string) {
	rewritten := make([]string, 0, len(args))

	var warnings []string

	for _, arg := range args {
		rewrittenArg, warning := rewriteDeprecatedArg(arg)
		rewritten = append(rewritten, rewrittenArg)

		if warning != "" {
			warnings = append(warnings, warning)
		}
	}

	return rewritten, warnings
}

func rewriteDeprecatedArg(arg string) (string, string) {
	switch arg {
	case "--noColor":
		return "--no-color", noColorDeprecation
	case "--noColor=true":
		return "--no-color=true", noColorDeprecation
	case "--noColor=false":
		return "--no-color=false", noColorDeprecation
	case "--json":
		return "--format=json", jsonDeprecation
	case "--json=true":
		return "--format=json", jsonDeprecation
	case "--json=false":
		return "--format=table", jsonDeprecation
	default:
		return arg, ""
	}
}

// printWarnings writes warnings to stderr. Write errors are not actionable
// (a closed terminal cannot be reported to), matching the renderer's
// deliberate ignore policy.
func printWarnings(w io.Writer, warnings []string) {
	for _, msg := range warnings {
		_, _ = fmt.Fprintf(w, warningLine, msg) //nolint:erraudit
	}
}
