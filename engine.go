package upd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Masterminds/semver/v3"
)

const (
	progressNameWidth = 24
	truncPrefix       = 19
	truncEllipsis     = "..."
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
		registry: NewRegistryClient(cfg),
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

		go func(pkgName string) {
			defer wg.Done()
			defer func() { <-sem }()

			pkg, bytes, err := e.registry.FetchPackument(ctx, strings.ToLower(pkgName))
			totalBytes.Add(int64(bytes))

			msg := truncMsg(pkgName)
			e.reporter.Tick(msg, int(totalBytes.Load()))

			mu.Lock()
			results[pkgName] = &FetchResult{name: pkgName, pkg: pkg, bytes: bytes, err: err}
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
) (int, int) {
	var updates, errors int

	for _, name := range manifest.SortedNames() {
		for _, spec := range manifest[name] {
			if spec.State != StateCheck {
				continue
			}

			updates, errors = e.applyOne(spec, results[name], pkg, updates, errors)
		}
	}

	return updates, errors
}

func (e *Engine) applyOne(
	spec *Spec,
	result *FetchResult,
	pkg *PackageFile,
	updates, errors int,
) (int, int) {
	if !resolveSpecVersion(spec, result, e.cfg) {
		errors++

		return updates, errors
	}

	if !shouldUpdate(spec) {
		return updates, errors
	}

	spec.State = StateUpdated
	updates++

	if e.cfg.Nop {
		return updates, errors
	}

	writeErr := pkg.UpdateDependency(spec.Section, spec.Name, spec.SNew)
	if writeErr != nil {
		spec.State = StateError
		spec.Err = fmt.Errorf("write %q in %q: %w", spec.Name, spec.Section, writeErr)
		errors++
	}

	return updates, errors
}

func resolveSpecVersion(spec *Spec, result *FetchResult, cfg *Config) bool {
	if result == nil {
		spec.State = StateError
		spec.Err = fmt.Errorf("%q: %w", spec.Name, ErrPackageNotFound)

		return false
	}

	if result.err != nil && result.pkg == nil {
		spec.State = StateError
		spec.Err = result.err

		return false
	}

	vNew, err := pickVersion(result.pkg, cfg)
	if err != nil {
		spec.State = StateError
		spec.Err = fmt.Errorf("resolve version for %q: %w", spec.Name, err)

		return false
	}

	spec.VNew = vNew
	if spec.IsLatest {
		spec.SNew = vNew
	} else {
		spec.SNew = replaceVersion(spec.SOld, spec.VOld, vNew)
	}

	return true
}

func pickVersion(pkg *Packument, cfg *Config) (string, error) {
	if cfg.Greatest {
		return pkg.GreatestVersion()
	}

	return pkg.LatestVersion()
}

func shouldUpdate(spec *Spec) bool {
	if spec.VOld == spec.VNew {
		spec.State = StateKept

		return false
	}

	if spec.IsLatest {
		return true
	}

	if !versionIsGreater(spec.VOld, spec.VNew) {
		spec.State = StateKept

		return false
	}

	return true
}

func versionIsGreater(oldVer, newVer string) bool {
	oldV, errOld := semver.NewVersion(oldVer)

	newV, errNew := semver.NewVersion(newVer)
	if errOld != nil || errNew != nil {
		return true
	}

	return newV.GreaterThan(oldV)
}

func replaceVersion(sOld, vOld, vNew string) string {
	return strings.Replace(sOld, vOld, vNew, 1)
}

func truncMsg(name string) string {
	if len(name) > progressNameWidth {
		return name[:truncPrefix] + truncEllipsis
	}

	return name + strings.Repeat(" ", progressNameWidth-len(name))
}

func (r FetchResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("%s: error: %v", r.name, r.err)
	}

	return fmt.Sprintf("%s: ok (%d bytes)", r.name, r.bytes)
}
