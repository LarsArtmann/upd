package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/LarsArtmann/upd/internal"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\x1b[31mERROR:\x1b[0m %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := internal.ParseFlags(args)
	if err != nil {
		if errors.Is(err, internal.ErrHelp) || errors.Is(err, internal.ErrVersion) {
			return nil
		}

		return err
	}

	pkg, err := internal.ReadPackageFile(cfg.File)
	if err != nil {
		return err
	}

	// Honor embedded "upd" field in package.json
	if embedded := pkg.GetUpdArgs(); len(embedded) > 0 {
		cfg.Patterns = append(embedded, cfg.Patterns...)
	}

	manifest := internal.BuildManifest(pkg, cfg.Patterns)
	toCheck := manifest.ToCheck()

	engine := internal.NewEngine(cfg)

	if !cfg.Quiet && len(toCheck) > 0 {
		reporter := internal.NewProgressReporter(os.Stderr, len(toCheck), cfg.NoColor)
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
	cfg *internal.Config,
	manifest internal.Manifest,
	pkg *internal.PackageFile,
	updates, errCount int,
) error {
	if !cfg.Quiet {
		renderer := internal.NewRenderer(os.Stdout, cfg.NoColor)
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
