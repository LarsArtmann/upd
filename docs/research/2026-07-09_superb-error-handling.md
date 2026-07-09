# Superb Error Handling for `upd` — Design & Implementation

**Date:** 2026-07-09
**Status:** Implemented across two passes, verified (build + vet + race + golangci-lint, 0 issues)
**Companion to:** `2026-07-09_error-handling-libraries.md` (the library-adoption report)

---

## TL;DR

We are **not** adopting go-error-family / bridge / samber-oops — none solve a problem `upd` has today. Instead we fixed the **nine concrete gaps** found across two audit passes, using only the standard library. No new dependencies. Every error now carries its reason to the user, registry outages are distinguished from typos, malformed sections and invalid patterns produce visible warnings instead of silent skips, and transient failures exit with code 75 so CI can retry.

---

## 1. The audit — nine gaps in the status quo

A full read of every error site (`errors.go`, `npm.go`, `packagejson.go`, `engine.go`, `config.go`, `manifest.go`, `render.go`, `cmd/upd/main.go`) across two passes found these weaknesses, ordered by impact:

| #   | Gap                                                                                                                                                                                                                             | Where                  | Severity                      |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | ----------------------------- |
| 1   | **Per-package errors were invisible.** `FetchResult.err` was captured but its message discarded in `resolveSpecVersion`; write errors in `applyOne` were counted and dropped. The user saw "2 errors" with zero clue _why_.     | `engine.go`            | **Critical**                  |
| 2   | **Registry errors were misclassified.** Every non-200 response (including 500/502/503 outages) was wrapped as `ErrPackageNotFound`. A typo and a registry outage produced the same error.                                       | `npm.go:59`            | **High**                      |
| 3   | **One exit code for everything.** `os.Exit(1)` hardcoded in `main()`. CI could not distinguish "fix your config" from "the registry was down, retry".                                                                           | `cmd/upd/main.go`      | Medium                        |
| 4   | **Two identical wrapping messages.** Request-build failure and request-send failure both said `"package information retrieval failed"` — impossible to tell which step broke.                                                   | `npm.go:46,54`         | Medium                        |
| 5   | **`GetDependencySection` silently returned empty on malformed sections.** If `"dependencies": 42` (number instead of object), the tool returned `{}` and reported "all up-to-date" — skipping all dependencies without warning. | `packagejson.go:51,58` | **High**                      |
| 6   | **`GetUpdArgs` silently returned nil on parse failure.** Malformed embedded `"upd"` config was invisibly ignored — the user's intended patterns never applied.                                                                  | `packagejson.go:72`    | Medium                        |
| 7   | **Invalid glob patterns silently dropped.** `splitPatterns` used `continue` on `glob.Compile` errors — the user's filter was invisibly ignored.                                                                                 | `manifest.go:174`      | Medium                        |
| 8   | **No structured error rendering.** The terminal table showed `error` as a state label but no diagnostic detail.                                                                                                                 | `render.go`            | Medium                        |
| 9   | **Silently swallowed parse errors.** `VersionKeys()` returns `nil` on unmarshal failure with no signal; `GreatestVersion` skips unparseable versions silently.                                                                  | `npm.go:122,99`        | Low (best-effort, acceptable) |

---

## 2. The design — what "superb" means for a stdlib-only CLI

Without a library, superb error handling rests on five idiomatic-Go pillars:

1. **Every error carries its reason to the boundary.** No error is counted and dropped — it is attached to the spec it concerns and rendered to the user.
2. **Sentinels model behavior, not just identity.** `ErrPackageNotFound` (user typo) vs `ErrRegistryUnavailable` (transient outage) drive different exit codes.
3. **Exit codes are meaningful.** BSD sysexits.h convention: `75` (EX_TEMPFAIL) for transient failures, `1` for everything else. CI scripts can `&& retry` on 75.
4. **Wrapping messages name the operation.** `"build registry request"`, `"send registry request"`, `"resolve version for %q"`, `"write %q in %q"` — each call site is distinguishable from the error string alone.
5. **`errors.Is` remains the only matching idiom** (no `errors.As` needed yet) — sentinels are the contract.

---

## 3. What changed (file by file)

### `errors.go` — new sentinel

```go
ErrRegistryUnavailable = errors.New("NPM registry is unavailable")
```

Splits the old single `ErrPackageNotFound` into two behavioral categories: not-found (user fault) vs unavailable (system fault, retryable).

### `manifest.go` — `Spec.Err error`

Added an `Err` field to `Spec`. Every spec that ends in `StateError` now carries the concrete error that caused it. This is the single highest-impact change: it is what makes gap #1 fixable.

### `npm.go` — accurate classification + distinct messages

- `classifyRegistryError(status, name)` helper: `404`/`410` → `ErrPackageNotFound`; everything else → `ErrRegistryUnavailable`. Both still wrap the status code and package name for diagnostics.
- Request-build failure: `"build registry request for %q"` (was: generic "package information retrieval failed").
- Request-send failure: `"send registry request for %q"` (was: identical generic message).

### `engine.go` — no more swallowed errors

Three error paths in `resolveSpecVersion` + `applyOne` that previously set `StateError` and discarded the cause now populate `spec.Err`:

| Path                        | `spec.Err` value                                         |
| --------------------------- | -------------------------------------------------------- |
| fetch failed (no packument) | the raw fetch error (already contextualized by `npm.go`) |
| version resolution failed   | `fmt.Errorf("resolve version for %q: %w", name, err)`    |
| byte-splice write failed    | `fmt.Errorf("write %q in %q: %w", name, section, err)`   |

### `render.go` — error detail block

After the upgrade table, a new `renderErrorDetails(manifest)` block lists every errored package with its reason:

```
Errors (2):
  brokenpkg           registry returned status 404 for "brokenpkg": package not found in NPM registry
  corruptdep          resolve version for "corruptdep": no "latest" dist-tag found
```

The table layout is untouched; the detail block is purely additive output below it.

### `cmd/upd/main.go` — exit codes

```go
const exitTransient = 75

func exitCode(err error) int {
    if errors.Is(err, upd.ErrRegistryUnavailable) {
        return exitTransient
    }
    return 1
}
```

`main()` now calls `os.Exit(exitCode(err))` instead of hardcoded `os.Exit(1)`. A registry outage during a fatal-path operation exits 75; everything else exits 1.

### `packagejson.go` — no more silent swallowing (pass 2)

**`GetDependencySection`** signature changed from `map[string]string` to `(map[string]string, error)`. Two error paths that previously returned `make(map[string]string)` (silent empty) now return errors:

- Top-level unmarshal failure → `fmt.Errorf("parse top-level JSON for section %q: %w", ...)`
- Section exists but wrong type (e.g. `"dependencies": 42`) → `fmt.Errorf("section %q is %s, expected object: %w", ..., ErrInvalidJSON)`
- Section unmarshal to `map[string]string` fails → `fmt.Errorf("parse section %q: expected object of name→version strings: %w", ...)`
- **Missing** section → still returns `(empty map, nil)` — absence is not an error

**`GetUpdArgs`** signature changed from `[]string` to `([]string, error)`. Parse failures now return errors instead of nil. Missing `upd` field still returns `(nil, nil)` — absence is not an error.

### `manifest.go` — warnings pipeline (pass 2)

**`BuildManifest`** signature changed from `Manifest` to `(Manifest, []string)` where the second return is a slice of warning strings. Sources:

- `GetDependencySection` errors → `"section %q: <error>"`
- Invalid glob patterns from `compilePatterns` → `"invalid glob pattern %q: <error>"`

Malformed sections produce a warning but the manifest still processes other valid sections. This is the right call: one broken section shouldn't abort the entire run.

**`splitPatterns` → `compilePatterns`**: renamed and restructured. Returns `(compiledPatterns, []string)` — invalid globs produce a warning instead of silent `continue`. A new `compiledPatterns` struct holds positive/negative compiled globs, and `matchesPatterns` takes pre-compiled patterns instead of recompiling on every call.

### `cmd/upd/main.go` — warning output (pass 2)

`main()` now:

1. Handles the error from `GetUpdArgs()` (fatal — malformed embedded config is a user error)
2. Prints `BuildManifest` warnings to stderr as `WARNING: <message>` (yellow)
3. Continues execution — warnings are informational, not fatal

---

## 4. Tests added

| Test                                           | File               | Asserts                                                                    |
| ---------------------------------------------- | ------------------ | -------------------------------------------------------------------------- |
| `TestRegistryClassifiesNotFoundAsRejection`    | `engine_test.go`   | 404 wraps `ErrPackageNotFound`, **not** `ErrRegistryUnavailable`           |
| `TestRegistryClassifiesServerErrorAsTransient` | `engine_test.go`   | 500/502/503 all wrap `ErrRegistryUnavailable`                              |
| `TestApplyUpdatesPopulatesSpecErr`             | `engine_test.go`   | a 404 sets `spec.Err` and it wraps `ErrPackageNotFound`                    |
| `TestRenderTableErrorDetailSurfacesReason`     | `render_test.go`   | error block shows header, package name, and reason message                 |
| `TestBuildManifestWarnsOnMalformedSection`     | `manifest_test.go` | `"devDependencies": 42` produces a warning; valid sections still processed |
| `TestCompilePatternsWarnsOnInvalidGlob`        | `manifest_test.go` | `[invalid` glob produces a warning about the pattern                       |

All existing tests pass (updated for new signatures: `BuildManifest` → `(Manifest, []string)`, `GetUpdArgs` → `([]string, error)`, `GetDependencySection` → `(map, error)`).

---

## 5. Verification

```
GOEXPERIMENT=jsonv2 go build ./...        # clean
GOEXPERIMENT=jsonv2 go vet ./...          # clean
GOEXPERIMENT=jsonv2 go test -race ./...   # ok, 1.026s
golangci-lint run ./...                   # 0 issues
```

---

## 6. Deliberate non-changes (judgment calls)

These were considered and **not** implemented — documented so the reasoning is visible:

1. **Exit non-zero when per-package errors occur but the run succeeds.** Currently a run that updates 3 packages and fails 2 still writes the file and exits 0 (errors are "soft", shown in the table). Changing this would alter exit-code semantics that CI scripts may depend on. **Recommend:** add an `--fail-on-error` flag if users request it, rather than changing the default.

2. **`errors.As` for typed network-error extraction.** The exit-code logic currently keys off `ErrRegistryUnavailable` (a sentinel), which is sufficient. If finer granularity is ever needed (e.g. DNS failure vs TCP timeout vs TLS error), `errors.As(err, &net.OpError{})` can be added to `exitCode()` without restructuring.

3. **Silent parse-error paths in `VersionKeys` / `GreatestVersion`.** These return empty/continue on unmarshal failure — defensible as best-effort parsing over registry data we don't control. Left as-is; `GreatestVersion` still returns `ErrNoSemverVersions` when _all_ versions are unparseable, which is the meaningful signal.

4. **Structured logging / observability.** `upd` has no logging stack and none is planned. If one is added, the `spec.Err` values are already structured enough to feed `slog` directly.

---

## 7. When to revisit the library question

The library report's revisit triggers still stand. This stdlib work does not close the door on go-error-family/oops — it raises the floor so that **if** `upd` later grows a retry loop or HTTP API, the error vocabulary is already behavioral (sentinels model retry-vs-not), and adopting a library would be an enrichment rather than a rescue.
