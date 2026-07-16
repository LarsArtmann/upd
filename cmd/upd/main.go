package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"charm.land/fang/v2"
	"github.com/LarsArtmann/upd"
	errorfamily "github.com/larsartmann/go-error-family"
)

func main() {
	err := run()
	if err != nil {
		os.Exit(errorfamily.ExitCode(err))
	}
}

func run() error {
	cmd, cfg := upd.NewCommand(executeRun)
	cmd.Version = upd.ProgramVersion
	cmd.SetVersionTemplate(versionTemplate)

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

func executeRun(ctx context.Context, cfg *upd.Config) error {
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
