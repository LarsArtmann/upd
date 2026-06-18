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
		os.Exit(1)
	}
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
	if embedded := pkg.GetUpdArgs(); len(embedded) > 0 {
		cfg.Patterns = append(embedded, cfg.Patterns...)
	}

	manifest := upd.BuildManifest(pkg, cfg.Patterns)
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
			return fmt.Errorf("failed to write package configuration file %q: %w", cfg.File, err)
		}
	}

	return nil
}
