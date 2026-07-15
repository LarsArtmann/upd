package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/LarsArtmann/upd"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	err := run(os.Args[1:])
	if err != nil {
		os.Exit(errorfamily.HandleError(err))
	}
}

func run(args []string) error {
	cfg, err := upd.ParseFlags(args)
	if err != nil {
		if errors.Is(err, upd.ErrHelp) || errors.Is(err, upd.ErrVersion) {
			return nil
		}

		return err
	}

	if !cfg.NoColor {
		cfg.NoColor = upd.ShouldDisableColor(os.Stdout)
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

	manifest, warnings := upd.BuildManifest(pkg, cfg.Patterns, cfg.PinLatest)

	if !cfg.Quiet {
		printWarnings(os.Stderr, warnings)
	}

	toCheck := manifest.ToCheck()
	engine := upd.NewEngine(cfg)

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
				return err
			}
		} else {
			renderer := upd.NewRenderer(os.Stdout, upd.RendererOptions{NoColor: cfg.NoColor, Verbose: cfg.Verbose})
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

func printWarnings(w *os.File, warnings []string) {
	for _, msg := range warnings {
		_, _ = fmt.Fprintf(w, "\x1b[33mWARNING:\x1b[0m %s\n", msg)
	}
}
