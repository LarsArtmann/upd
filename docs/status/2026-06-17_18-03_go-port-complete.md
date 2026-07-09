# Status Report: UPD Go Port

**Date:** 2026-06-17 18:03
**Branch:** master
**Commits ahead of origin:** 6

> **Current Status (reviewed 2026-07-09):** Historical snapshot of the initial
> JS→Go port. All items in section (a) FULLY DONE remain accurate.
>
> **Completed since this report:**
>
> - `flake.nix` — DONE (build/test/lint/run/demo apps)
> - GitHub Actions CI — DONE (`.github/workflows/ci.yml`)
> - `AGENTS.md` — DONE (comprehensive project context)
> - Version injection via ldflags — DONE (`ProgramVersion` var, set in flake + CI)
> - Atomic writes with TOCTOU protection — DONE (`go-atomic-write` v0.2.0)
> - `--pinLatest` / `-P` flag — DONE
> - `encoding/json/v2` migration — DONE (replaced `tidwall/gjson`)
>
> **Rejected / Obsolete:**
>
> - `internal/` package layout — **OBSOLETE**: code was restructured to the repo
>   root as package `upd`; file paths listed in the File Inventory below are
>   historically accurate but no longer match the repo.
> - Docker — **REJECTED**: no Dockerfile; single static binary makes it unnecessary.
> - Update notifier — **REJECTED**: Go binaries don't self-update.
> - Shell completions — not done (YAGNI for current scope).
>
> **Still relevant:**
>
> - gosec / govulncheck — still not run.
> - Bench tests — still not written.
> - The quiet/non-quiet code duplication in `main.go` — still present.

---

## Executive Summary

The `upd` CLI — a tool for upgrading NPM package dependencies in `package.json`
while preserving JSON formatting — has been **fully ported from JavaScript to Go**.

The original was 414 lines of JS across 1 file with 15 npm dependencies.
The Go port is ~900 lines across 9 source files with **3 Go dependencies**.

Build passes, vet passes, 22 tests pass, and the binary was verified end-to-end
against the live NPM registry (real package lookups, real writes, formatting
preservation confirmed via diff).

---

## a) FULLY DONE

| Area                             | Details                                                                                                                                                                        | Verified                   |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | -------------------------- |
| **Project scaffold**             | `go.mod`, `cmd/upd/`, `internal/`, `.gitignore`                                                                                                                                | ✅ `go build ./...`        |
| **CLI flag parsing**             | All original flags: `-h -V -q -n -C -f -g -a -c`, short+long forms, positional patterns                                                                                        | ✅ Manual test             |
| **Version info**                 | `-V` prints name, version, URL, description, copyright                                                                                                                         | ✅ Manual test             |
| **Help output**                  | `-h` prints full usage with flag descriptions                                                                                                                                  | ✅ Manual test             |
| **package.json reader**          | Reads file, validates JSON, returns `*PackageFile` with raw bytes                                                                                                              | ✅ Unit test               |
| **Formatting-preserving writer** | `UpdateDependency` uses gjson byte offsets to replace only the value — indentation, quotes, key ordering, trailing newlines all preserved                                      | ✅ Unit test + manual diff |
| **Embedded args (`upd` field)**  | Reads `upd` field from package.json (string or array), prepends to CLI patterns                                                                                                | ✅ Unit test               |
| **Manifest builder**             | Extracts deps from 4 sections (`optionalDependencies`, `peerDependencies`, `devDependencies`, `dependencies`), applies glob filtering, extracts semver from constraint strings | ✅ Unit test               |
| **Glob pattern matching**        | Positive patterns, negative (`!` prefix) exclusion, scoped packages (`@scope/*`)                                                                                               | ✅ Unit test (10 cases)    |
| **Version extraction**           | Regex extracts version from `^`, `~`, bare version strings; skips `file:`, `git:`, `>=` ranges                                                                                 | ✅ Unit test (10 cases)    |
| **NPM registry client**          | HTTP GET to registry.npmjs.org, 20s timeout, User-Agent header, reads full packument                                                                                           | ✅ Manual test             |
| **Latest version resolution**    | Reads `dist-tags.latest` from packument                                                                                                                                        | ✅ Unit test               |
| **Greatest version resolution**  | Iterates `versions` keys, sorts via semver, returns highest                                                                                                                    | ✅ Unit test               |
| **Concurrent fetch engine**      | Semaphore-based concurrency limit (default 8), atomic byte counter, `sync.WaitGroup` coordination                                                                              | ✅ Manual test             |
| **State machine**                | `todo → check → kept/updated/error/skipped/ignored` with semver comparison                                                                                                     | ✅ Manual test             |
| **Semver comparison**            | Uses `Masterminds/semver/v3` for `GreaterThan` checks; won't downgrade                                                                                                         | ✅ Unit test               |
| **Character-level diff**         | LCS-based algorithm for highlighting version changes in table output                                                                                                           | ✅ Unit test (3 cases)     |
| **Colored table renderer**       | Unicode box-drawing borders, ANSI color codes, `--noColor` support, bold headers                                                                                               | ✅ Manual test             |
| **Diff highlighting**            | Old version: red diff; New version: green diff (matching original JS behavior)                                                                                                 | ✅ Manual test             |
| **Progress bar**                 | Unicode progress bar on stderr during concurrent fetch, clears on completion                                                                                                   | ✅ Manual test             |
| **Error handling**               | Sentinel errors (`ErrFileNotFound`, `ErrInvalidJSON`, `ErrPackageNotFound`), wrapped with context, red `ERROR:` prefix                                                         | ✅ Manual test             |
| **Write-back**                   | Writes updated `package.json` only when updates occurred and `-n` not set                                                                                                      | ✅ Manual diff             |
| **JS file cleanup**              | Removed: `upd.js`, `package.json`, `eslint.mjs`, `Makefile`, `Dockerfile`, `.dockerignore`, `.npmignore`, `screenshot.png`                                                     | ✅ `git rm`                |
| **README**                       | Updated with Go install instructions, usage, development commands                                                                                                              | ✅ Written                 |
| **Test suite**                   | 22 tests across 3 test files covering: regex, glob, manifest, formatting, diff, semver                                                                                         | ✅ `go test ./...`         |
| **Production build**             | `go build -ldflags="-s -w" -trimpath` → 6.6 MB binary                                                                                                                          | ✅ Built                   |

---

## b) PARTIALLY DONE

| Area                    | Status                           | Gap                                                                                                                                                                                                                                 |
| ----------------------- | -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **LSP warnings**        | Build/vet/test all pass clean    | golangci*lint_ls reports 3 stale-cache warnings (errcheck on `resp.Body.Close`, typecheck on `config.go` `PrintUsage`) — these are false positives from LSP cache not catching the `defer func() { * = ... }()`and`io.Writer` fixes |
| **Progress bar polish** | Works but overwrites with spaces | The `Finish()` method uses a fixed 80-char clear which may not match terminal width on all terminals; original JS used the `progress` library which handled this                                                                    |

---

## c) NOT STARTED

| Area                                 | Notes                                                                                                                                        |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **`-g` (greatest) integration test** | Unit tested `GreatestVersion()` but never ran the full CLI with `-g` against live registry                                                   |
| **Docker**                           | Original had a Dockerfile; no Go-based Dockerfile created                                                                                    |
| **CI/CD**                            | No GitHub Actions workflow                                                                                                                   |
| **`flake.nix`**                      | No Nix flake for development/build (LarsArtmann projects use `flake.nix` per AGENTS.md)                                                      |
| **gosec / govulncheck**              | No security scanning run                                                                                                                     |
| ** Goreleaser / release automation** | No release pipeline                                                                                                                          |
| **Shell completions**                | No bash/zsh/fish completion generation                                                                                                       |
| **Update notifier**                  | Original JS had `update-notifier` to check for newer `upd` versions; not ported (intentionally — Go binaries don't self-update the same way) |
| **`AGENTS.md`**                      | No project-specific AGENTS.md written yet                                                                                                    |
| **`FEATURES.md` / `TODO_LIST.md`**   | Not created                                                                                                                                  |

---

## d) TOTALLY FUCKED UP

**Nothing.** No broken code, no failing tests, no data loss, no corrupted state.

The only notable bug during development was the **diff coalescing bug** (chunk
concatenation was prepending instead of appending characters, producing reversed
diff output) — caught and fixed by tests before any commit.

---

## e) WHAT WE SHOULD IMPROVE

### High Priority

1. **`min`/`max` builtins** — `progress.go` uses custom `if` logic instead of Go 1.21+ `min()` builtin; gopls hints at this
2. **Tagged switch in renderer** — `render.go:104` uses `switch` with cases on booleans instead of tagged switch on `spec.State`; gopls hints
3. **`strings.Contains` everywhere** — test file had custom `contains`/`indexOf` helpers; cleaned up but should audit all files
4. **No integration test** — No test that exercises the full pipeline (read → fetch → compare → write) with a mock registry
5. **No bench tests** — No performance benchmarks for diff, glob, or manifest building
6. **Duplicated main.go logic** — `run()` has two branches (quiet/non-quiet) that duplicate `FetchAll` + `ApplyUpdates` + `RenderTable` + `Write`; should extract a shared helper

### Medium Priority

7. **Progress bar width** — Hardcoded 80-char clear; should detect terminal width or use `\r` + overwrite
8. **HTTP client reuse** — `FetchPackument` creates a new `http.Client` per call; should share one client across all goroutines for connection pooling
9. **Context timeout** — No overall deadline on the fetch phase; individual requests have 20s but 1000 deps × retries could take long
10. **Retry logic** — No retries on transient registry errors (429, 5xx)
11. **Registry config** — Hardcoded `registry.npmjs.org`; original used pacote which respects `.npmrc`
12. **Scoped package encoding** — `url.PathEscape` may not match NPM's expected encoding for scoped packages (`@scope/name` → `@scope%2Fname`)

### Low Priority

13. **Color detection** — `--noColor` is manual flag; should also auto-detect non-TTY (pipe) and disable colors
14. **Table column widths** — Hardcoded to match original JS; long package names (>37 chars) get truncated but not elegantly
15. **Structured logging** — No `slog` logging; errors go to stderr only
16. **Version string injection** — `ProgramVersion` is hardcoded `1.0.0`; should use `ldflags` injection from git tags

---

## f) Top 25 Things to Get Done Next

| #   | Task                                                        | Impact | Effort | Category       |
| --- | ----------------------------------------------------------- | ------ | ------ | -------------- |
| 1   | Fix gopls hints: use `min()` builtin, tagged switch         | Low    | 5 min  | Code quality   |
| 2   | Extract shared helper in `main.go` to remove duplication    | Medium | 15 min | Code quality   |
| 3   | Add integration test with mock HTTP registry server         | High   | 1 hour | Testing        |
| 4   | Share `http.Client` across goroutines in engine             | Medium | 15 min | Performance    |
| 5   | Add `-g` (greatest) end-to-end manual test                  | Low    | 5 min  | Testing        |
| 6   | Write `flake.nix` for dev/build/test/lint                   | High   | 30 min | DevOps         |
| 7   | Write project `AGENTS.md` with architecture decisions       | High   | 20 min | Documentation  |
| 8   | Auto-detect non-TTY and disable colors                      | Medium | 10 min | UX             |
| 9   | Add retry logic for transient registry errors (429, 5xx)    | Medium | 30 min | Reliability    |
| 10  | Add overall context timeout for fetch phase                 | Medium | 15 min | Reliability    |
| 11  | Add bench tests for diff and glob                           | Low    | 20 min | Testing        |
| 12  | Inject version via `-ldflags` from git tag                  | Low    | 10 min | Release        |
| 13  | Create GitHub Actions CI (build, test, vet, lint)           | High   | 30 min | DevOps         |
| 14  | Run `gosec` and `govulncheck` and fix findings              | Medium | 20 min | Security       |
| 15  | Write `Dockerfile` for Go (multi-stage, scratch/distroless) | Medium | 20 min | DevOps         |
| 16  | Verify scoped package URL encoding against live registry    | Medium | 15 min | Correctness    |
| 16  | Add `.npmrc` parsing for custom registry support            | Medium | 30 min | Feature parity |
| 18  | Add `FEATURES.md` with feature inventory                    | Low    | 15 min | Documentation  |
| 15  | Add `TODO_LIST.md` with short-term tasks                    | Low    | 15 min | Documentation  |
| 20  | Generate shell completions (bash/zsh/fish)                  | Low    | 20 min | UX             |
| 21  | Add `--registry <url>` flag for custom registry             | Medium | 15 min | Feature parity |
| 22  | Add JSON output mode (`--json`) for CI/scripting            | Medium | 30 min | Feature        |
| 23  | Add `--dry-run` as alias for `--nop`                        | Low    | 5 min  | UX             |
| 24  | Add timeout flag (`--timeout <seconds>`)                    | Low    | 10 min | UX             |
| 25  | Add Go module vulnerabilities badge to README               | Low    | 5 min  | Documentation  |

---

## g) Top Question I Cannot Figure Out Myself

**Should we keep this as a faithful 1:1 port of the original `upd`, or evolve it into
a more general dependency updater?**

The original is NPM-specific. The Go architecture (manifest builder, glob filter,
semver comparison, formatting-preserving writer) is general enough to support other
ecosystems (Go modules, Cargo, pip, etc.). The `Packument` / NPM registry client
is the only NPM-specific piece.

**Options:**

- **A) Faithful port** — Keep it NPM-only, match original behavior exactly, ship it
- **B) Plugin architecture** — Abstract the registry client and manifest format behind
  interfaces, support multiple ecosystems via plugins

Option A is faster to ship. Option B is more ambitious but the architecture is
already shaped for it. This is a product direction decision I can't make alone.

---

## Metrics

| Metric            | Original (JS)     | Port (Go)                            |
| ----------------- | ----------------- | ------------------------------------ |
| Source files      | 1 (`upd.js`)      | 9 (`.go` source) + 3 test files      |
| Source lines      | 414               | ~900                                 |
| Dependencies      | 15 npm packages   | 3 Go modules                         |
| Binary size       | N/A (runtime)     | 6.6 MB (static, stripped)            |
| Startup time      | ~300ms (Node.js)  | ~2ms (native)                        |
| Test count        | 0                 | 22                                   |
| Concurrency model | `awaity.mapLimit` | `sync.WaitGroup` + semaphore channel |

---

## File Inventory

```
cmd/upd/main.go              # Entry point: parse → read → fetch → apply → render → write
internal/config.go           # CLI flag definitions, help/version output
internal/diff.go             # LCS-based character-level diff algorithm
internal/diff_test.go        # Diff tests (equal, replace, insert)
internal/engine.go           # Concurrent fetch engine + state machine
internal/errors.go           # Sentinel error definitions
internal/manifest.go         # Dependency extraction, glob filtering, version regex
internal/manifest_test.go    # Manifest, regex, glob tests
internal/npm.go              # NPM registry HTTP client, version resolution
internal/packagejson.go      # Formatting-preserving JSON reader/writer
internal/packagejson_test.go # UpdateDependency, GetUpdArgs tests
internal/progress.go         # Progress bar reporter
internal/render.go           # Colored table renderer with diff highlighting
```
