package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/LarsArtmann/upd"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31mERROR:\x1b[0m %v\n", err)
		os.Exit(exitCode(err))
	}
}

const exitTransient = 75

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	if errors.Is(err, upd.ErrRegistryUnavailable) {
		return exitTransient
	}

	return 1
}

func run(args []string) error {
	cfg, err := upd.ParseFlags(args)
	if err != nil {
		if errors.Is(err, upd.ErrHelp) || errors.Is(err, upd.ErrVersion) {
			return nil
		}

		return fmt.Errorf("parse flags: %w", err)
	}

	// Auto-detect color suppression: NO_COLOR env var or non-TTY stdout
	if !cfg.NoColor {
		cfg.NoColor = upd.ShouldDisableColor(os.Stdout)
	}

	pkg, err := upd.ReadPackageFile(cfg.File)
	if err != nil {
		return fmt.Errorf("read package file: %w", err)
	}

	// Honor embedded "upd" field in package.json
	embedded, err := pkg.GetUpdArgs()
	if err != nil {
		return fmt.Errorf("read embedded upd args: %w", err)
	}

	if len(embedded) > 0 {
		cfg.Patterns = append(embedded, cfg.Patterns...)
	}

	manifest, warnings := upd.BuildManifest(pkg, cfg.Patterns, cfg.PinLatest)

	// In quiet mode, suppress warnings too; in JSON mode, emit warnings as JSON
	// metadata on stderr so stdout stays pure JSON
	if !cfg.Quiet {
		printWarnings(os.Stderr, warnings)
	}

	toCheck := manifest.ToCheck()
	engine := upd.NewEngine(cfg)

	// Progress bar only in interactive mode with work to do
	showProgress := !cfg.Quiet && len(toCheck) > 0

	reporter := upd.NewProgressReporter(os.Stderr, len(toCheck), cfg.NoColor)
	if showProgress {
		reporter.Start()
		engine = engine.WithReporter(reporter)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	results := engine.FetchAll(ctx, toCheck)

	if showProgress {
		reporter.Finish()
	}

	updates, errCount := engine.ApplyUpdates(manifest, results, pkg)

	return finalizeRun(cfg, manifest, pkg, updates, errCount)
}

func finalizeRun(
	cfg *upd.Config,
	manifest upd.Manifest,
	pkg *upd.PackageFile,
	updates, errCount int,
) error {
	if !cfg.Quiet {
		if cfg.JSON {
			err := upd.RenderJSON(os.Stdout, manifest, updates)
			if err != nil {
				return fmt.Errorf("render JSON output: %w", err)
			}
		} else {
			renderer := upd.NewRenderer(os.Stdout, upd.RendererOptions{NoColor: cfg.NoColor, Verbose: cfg.Verbose})
			renderer.RenderTable(manifest, updates, errCount, cfg.All)
		}
	}

	if updates > 0 && !cfg.Nop {
		err := pkg.Write(cfg.File)
		if err != nil {
			if errors.Is(err, upd.ErrConcurrentModification) {
				return fmt.Errorf("%w; your file was not changed — re-run upd", err)
			}

			return fmt.Errorf("write package file: %w", err)
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d package(s) failed: %w", errCount, upd.ErrPartialFailure)
	}

	return nil
}

func printWarnings(w *os.File, warnings []string) {
	for _, msg := range warnings {
		_, _ = fmt.Fprintf(w, "\x1b[33mWARNING:\x1b[0m %s\n", msg)
	}
}
