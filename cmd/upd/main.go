package main

import (
	"context"
	"errors"
	"fmt"
	"os"

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
	printWarnings(os.Stderr, warnings)

	toCheck := manifest.ToCheck()

	engine := upd.NewEngine(cfg)

	if !cfg.Quiet && len(toCheck) > 0 {
		reporter := upd.NewProgressReporter(os.Stderr, len(toCheck), cfg.NoColor)
		reporter.Start()
		engine = engine.WithReporter(reporter)
		results := engine.FetchAll(context.Background(), toCheck)

		reporter.Finish()

		updates, errCount := engine.ApplyUpdates(manifest, results, pkg)

		return finalizeRun(cfg, manifest, pkg, updates, errCount)
	}

	// Quiet mode or nothing to check: no progress bar
	results := engine.FetchAll(context.Background(), toCheck)
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
		renderer := upd.NewRenderer(os.Stdout, cfg.NoColor)
		renderer.RenderTable(manifest, updates, errCount, cfg.All)
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
