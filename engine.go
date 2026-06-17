package upd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"
)

type Reporter interface {
	Tick(msg string, bytes int)
}

type noopReporter struct{}

func (noopReporter) Tick(string, int) {}

type Engine struct {
	cfg      *Config
	registry *RegistryClient
	reporter Reporter
}

func NewEngine(cfg *Config) *Engine {
	return &Engine{
		cfg:      cfg,
		registry: NewRegistryClient(cfg.UserAgent()),
		reporter: noopReporter{},
	}
}

func (e *Engine) WithReporter(r Reporter) *Engine {
	e.reporter = r

	return e
}

// FetchResult holds the outcome of fetching a single package's packument
// from the NPM registry. Callers receive a map of these from FetchAll and
// pass it opaquely to ApplyUpdates.
type FetchResult struct {
	name  string
	pkg   *Packument
	bytes int
	err   error
}

func (e *Engine) FetchAll(ctx context.Context, names []string) map[string]*FetchResult {
	results := make(map[string]*FetchResult, len(names))

	var mu sync.Mutex

	sem := make(chan struct{}, e.cfg.Concurrency)

	var (
		wg         sync.WaitGroup
		totalBytes atomic.Int64
	)

	for _, name := range names {
		wg.Add(1)

		sem <- struct{}{}

		go func(n string) {
			defer wg.Done()
			defer func() { <-sem }()

			pkg, bytes, err := e.registry.FetchPackument(ctx, strings.ToLower(n))
			totalBytes.Add(int64(bytes))

			msg := truncMsg(n)
			e.reporter.Tick(msg, int(totalBytes.Load()))

			mu.Lock()
			results[n] = &FetchResult{name: n, pkg: pkg, bytes: bytes, err: err}
			mu.Unlock()
		}(name)
	}

	wg.Wait()

	return results
}

func (e *Engine) ApplyUpdates(
	manifest Manifest,
	results map[string]*FetchResult,
	pkg *PackageFile,
) (updates, errors int) {
	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if spec.State != StateCheck {
				continue
			}

			result, ok := results[name]
			if !ok || (result.err != nil && result.pkg == nil) {
				spec.State = StateError
				errors++

				continue
			}

			vNew, err := e.resolveVersion(result.pkg)
			if err != nil {
				spec.State = StateError
				errors++

				continue
			}

			spec.VNew = vNew
			spec.SNew = replaceVersion(spec.SOld, spec.VOld, vNew)

			if spec.VOld == spec.VNew {
				spec.State = StateKept

				continue
			}

			oldV, errOld := semver.NewVersion(spec.VOld)

			newV, errNew := semver.NewVersion(spec.VNew)
			if errOld == nil && errNew == nil && !newV.GreaterThan(oldV) {
				spec.State = StateKept

				continue
			}

			spec.State = StateUpdated
			updates++

			if !e.cfg.Nop {
				err := pkg.UpdateDependency(spec.Section, spec.Name, spec.SNew)
				if err != nil {
					spec.State = StateError
					errors++
				}
			}
		}
	}

	return updates, errors
}

func (e *Engine) resolveVersion(pkg *Packument) (string, error) {
	if e.cfg.Greatest {
		return pkg.GreatestVersion()
	}

	return pkg.LatestVersion()
}

func replaceVersion(sOld, vOld, vNew string) string {
	return strings.Replace(sOld, vOld, vNew, 1)
}

func truncMsg(name string) string {
	if len(name) > 24 {
		return name[:19] + "..."
	}

	return name + strings.Repeat(" ", 24-len(name))
}

func (r FetchResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("%s: error: %v", r.name, r.err)
	}

	return fmt.Sprintf("%s: ok (%d bytes)", r.name, r.bytes)
}
