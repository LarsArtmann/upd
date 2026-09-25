# Status Report: UPD Go Port — Round 2 Improvements

**Date:** 2026-06-17 18:26
**Branch:** master (pushed to origin)
**Commits since last report:** 7
**Test coverage:** 86.6% of statements
**Test assertions:** 67 (across 6 test files)

> **Current Status (reviewed 2026-07-09):** Historical snapshot. All items in
> section (a) FULLY DONE remain accurate. Test count has grown from 67 assertions
> to 51 test functions (across 7 test files). Coverage is higher.
>
> **Completed since this report:**
>
> - AGENTS.md — DONE (comprehensive, kept current)
> - Atomic writes with TOCTOU protection — DONE (`go-atomic-write` v0.2.0)
> - `--pinLatest` / `-P` flag — DONE
> - `encoding/json/v2` migration — DONE (replaced `tidwall/gjson`, down to 3 deps)
> - Error handling overhaul — DONE (`Spec.Err`, `ErrRegistryUnavailable`, exit
>   code 75 for transient failures, warnings pipeline, error detail block)
> - `makezero` lint warnings — RESOLVED (diff.go and render.go fixed)
>
> **Rejected / Obsolete:**
>
> - `internal/` layout — **OBSOLETE**: restructured to repo root as package `upd`.
>   File Inventory paths below are historically accurate but no longer match.
> - Dockerfile — **REJECTED**.
> - `Spec.Section` typed enum — **REJECTED**: bare string remains; YAGNI for current scope.
> - `PackageName` branded type — **REJECTED**: YAGNI.
> - `fetchResult` unexporting — **OBSOLETE**: renamed to `FetchResult` and documented.
>
> ~~**Still relevant (from the Top 25):**~~ all resolved since (docs-health pass 2026-09-25):
>
> - ~~golangci-lint in CI — CI still runs only `go vet` (#2)~~ done at `e64d3a7`
> - ~~Retry logic for 429/5xx — not done (#3)~~ done at `e64d3a7`
> - ~~Auto-detect non-TTY for color disabling — not done (#4)~~ done at `e64d3a7`
> - ~~`--registry` flag — not done (#5)~~ done at `e64d3a7`
> - ~~Context deadline for fetch phase — not done (#7)~~ done at `e64d3a7` (signal-aware context)
> - ~~govulncheck / gosec — not run (#10, #11)~~ done — govulncheck CI job; gosec via golangci config
> - ~~`--json` output mode — not done (#12)~~ done at `e64d3a7` (D34)

---

## Executive Summary

After the initial JS→Go port completion (previous report), this round focused on
**architecture quality, testability, and DevOps infrastructure**. The biggest wins
were extracting `RegistryClient` (enabling mock-based integration tests), adding
67 test assertions (up from 0 on the engine), and creating `flake.nix` + CI.

The project is now **production-ready** for its current feature scope.

---

## a) FULLY DONE

| Area                          | Details                                                                                                                                       | Verified                         |
| ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| **RegistryClient extraction** | Shared `*http.Client` with connection pooling; injectable for tests                                                                           | ✅ Build + 7 engine tests        |
| **Engine integration tests**  | 7 tests with `httptest.Server` mock registry: fetch success, apply updates, kept, nop mode, 404 error, write-back verification, greatest mode | ✅ All pass                      |
| **Render tests**              | 8 tests: updated table, all-up-to-date, all mode, error state, noColor ANSI stripping, color ANSI presence, visibleLength, centerPad          | ✅ All pass                      |
| **Config tests**              | 8 tests: defaults, short flags, long flags, multiple patterns, help, version, UserAgent                                                       | ✅ All pass                      |
| **86.6% test coverage**       | All business logic covered; only error-formatting branches in config.go uncovered                                                             | ✅ `go test -cover`              |
| **flake.nix**                 | `buildGoModule` package, devShell (go, gopls, gotools, staticcheck), apps: build/test/lint/run                                                | ✅ Written                       |
| **GitHub Actions CI**         | Build + vet + test on push/PR to master                                                                                                       | ✅ Written                       |
| **Version injection**         | `ProgramVersion` is a `var` (default "dev"), injectable via `-ldflags="-X ...=1.0.0"`                                                         | ✅ Verified binary shows `1.0.0` |
| **Dead code removal**         | Removed unused `contains`/`indexOf` helpers from engine_test.go                                                                               | ✅ Build clean                   |
| **Semantic cleanup**          | Moved `RegistryURL` from config.go to pnpm.go where it belongs                                                                                | ✅ Build clean                   |

## b) PARTIALLY DONE

| ~~**LSP warnings**~~        | ~~Build/vet/test all pass~~ done — errcheck fix landed; the stale-cache warnings cleared | ~~golangci_lint_ls reports `errcheck` warnings on `fmt.Fprintln`/`fmt.Fprintf` in config.go — these are intentional (writing to help/version output where errors are meaningless)~~ (historical: `PrintUsage` itself was later removed by the fang migration) |
| ~~**Progress bar**~~            | ~~Works functionally~~ done at `e64d3a7` — `COLUMNS` width detection (D42)      | ~~Hardcoded 80-char clear width; no terminal width detection~~ resolved (80 remains the fallback)                                                    |
| ~~**Scoped package encoding**~~ | ~~`url.PathEscape` used~~ done at `e64d3a7` — `TestScopedPackageURLEncoding` (mock)   | ~~Not verified against live scoped packages (`@scope/name`)~~ live check → TODO_LIST.md #26                                                       |

## c) NOT STARTED

| Area                     | Notes                                                       |
| ------------------------ | ----------------------------------------------------------- |
| ~~**Docker**~~               | ~~No Dockerfile for the Go version~~ Won't implement — R1                            |
| ~~**gosec / govulncheck**~~  | ~~No security scanning~~ done — govulncheck CI job + gosec linter                   |
| ~~**Shell completions**~~    | ~~No bash/zsh/fish completions~~ done at `81d8c44` (Cobra)                          |
| ~~**Retry logic**~~          | ~~No retries on transient registry errors (429, 5xx)~~ done at `e64d3a7`            |
| ~~**`.npmrc` support**~~     | ~~No custom registry URL from `.npmrc` → moved to TODO_LIST.md #12~~ |
| ~~**Auto color detection**~~ | ~~`--noColor` is manual only; doesn't auto-disable on non-TTY~~ done at `e64d3a7`   |
| ~~**JSON output mode**~~     | ~~No `--json` flag for CI/scripting~~ done at `e64d3a7`                             |
| ~~**Bench tests**~~          | ~~No performance benchmarks~~ done at `e64d3a7` (`benchmark_test.go`)               |

## d) TOTALLY FUCKED UP

**Nothing.** No broken code, no failing tests, no data loss.

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. ~~**No `AGENTS.md`** — Project-specific context for AI sessions not written yet~~ done (kept current since)
2. ~~**No retry/backoff** — Transient registry failures (429, 5xx) cause immediate error state~~ done at `e64d3a7`
3. ~~**No context deadline on fetch phase** — Individual requests have 20s timeout but no overall deadline~~ done at `e64d3a7` (signal-aware)
4. ~~**Progress bar terminal width** — Should detect terminal width instead of hardcoded 80~~ done at `e64d3a7` (D42)
5. ~~**HTTP client could use transport tuning** — MaxIdleConns, IdleConnTimeout for large dep lists~~ done at `e64d3a7`

### Testing

6. ~~**No bench tests** — Diff algorithm and glob matching have no performance regression detection~~ done at `e64d3a7` (CI run → ROADMAP theme 4)
7. ~~**No test for `-g` against live registry** — Only mock-tested~~ still open — folded into TODO_LIST.md #26
8. ~~**13.4% coverage gap** — Mostly error branches in config.go (PrintUsage/PrintVersion)~~ done differently — `PrintUsage`/`PrintVersion` removed by the fang migration (`81d8c44`); remaining coverage gap tracked via TODO_LIST.md #18

### DevOps

9. ~~**No release automation** — No GoReleaser or tag-based release pipeline~~ moved to TODO_LIST.md #3 — still open
10. ~~**No golangci-lint in CI** — CI runs `go vet` but not `golangci-lint`~~ done at `e64d3a7`
11. ~~**No Docker image** — Original JS had Docker; Go port doesn't~~ Won't implement — R1

### Type Model

12. ~~**`Spec.Section` is a bare string** — Could be a typed enum (`SectionDependencies`, etc.) for compile-time safety~~ Won't implement — R3
13. ~~**`fetchResult` exposed internally** — Could be unexported or replaced with a cleaner result type~~ done — renamed to `FetchResult` and documented (see header)
14. ~~**No branded types** — Package names are bare strings; a `PackageName` type would prevent confusion with version strings~~ Won't implement — R4

## f) Top 25 Things to Get Done Next

| #  | Task                                          | Impact | Effort | Category       |
| -- | --------------------------------------------- | ------ | ------ | -------------- |
| ~~1~~ | ~~Write project `AGENTS.md`~~ done                              | ~~High~~ | ~~20 min~~ | ~~Documentation~~ |
| ~~2~~ | ~~Add golangci-lint to CI~~ done at `e64d3a7`                   | ~~High~~ | ~~15 min~~ | ~~DevOps~~ |
| ~~3~~ | ~~Add retry logic for 429/5xx registry errors~~ done at `e64d3a7` | ~~High~~ | ~~30 min~~ | ~~Reliability~~ |
| ~~4~~ | ~~Auto-detect non-TTY and disable colors~~ done at `e64d3a7`    | ~~Medium~~ | ~~10 min~~ | ~~UX~~ |
| ~~5~~ | ~~Add `--registry <url>` flag~~ done at `e64d3a7`               | ~~Medium~~ | ~~15 min~~ | ~~Feature parity~~ |
| ~~6~~ | ~~Type `Section` as enum instead of bare string~~ Won't implement — R3 | ~~Medium~~ | ~~20 min~~ | ~~Type safety~~ |
| ~~7~~ | ~~Add context deadline for entire fetch phase~~ done at `e64d3a7` | ~~Medium~~ | ~~15 min~~ | ~~Reliability~~ |
| ~~8~~ | ~~Write Dockerfile (multi-stage, distroless)~~ Won't implement — R1 | ~~Medium~~ | ~~20 min~~ | ~~DevOps~~ |
| ~~9~~ | ~~Add bench tests for diff + glob~~ done at `e64d3a7`           | ~~Low~~ | ~~20 min~~ | ~~Testing~~ |
| ~~10~~ | ~~Run `govulncheck` and fix findings~~ done                     | ~~Medium~~ | ~~15 min~~ | ~~Security~~ |
| ~~11~~ | ~~Run `gosec` and fix findings~~ done (golangci config)         | ~~Medium~~ | ~~15 min~~ | ~~Security~~ |
| ~~12~~ | ~~Add `--json` output mode~~ done at `e64d3a7`                  | ~~Medium~~ | ~~30 min~~ | ~~Feature~~ |
| ~~13~~ | ~~Add GoReleaser config~~ moved to TODO_LIST.md #3             | ~~Medium~~ | ~~30 min~~ | ~~Release~~ |
| ~~14~~ | ~~Verify scoped package URL encoding live~~ done at `e64d3a7` (mock-verified; live check → TODO_LIST.md #26) | ~~Medium~~ | ~~15 min~~ | ~~Correctness~~ |
| ~~15~~ | ~~Add `--timeout` flag~~ done at `e64d3a7`                      | ~~Low~~ | ~~10 min~~ | ~~UX~~ |
| ~~16~~ | ~~Tune HTTP transport (MaxIdleConns, etc.)~~ done at `e64d3a7`  | ~~Low~~ | ~~10 min~~ | ~~Performance~~ |
| ~~17~~ | ~~Add `--dry-run` as alias for `--nop`~~ done at `e64d3a7`      | ~~Low~~ | ~~5 min~~ | ~~UX~~ |
| ~~18~~ | ~~Extract `PackageName` branded type~~ Won't implement — R4     | ~~Low~~ | ~~15 min~~ | ~~Type safety~~ |
| ~~19~~ | ~~Add `FEATURES.md`~~ done at `474f9dd`                         | ~~Low~~ | ~~15 min~~ | ~~Documentation~~ |
| ~~20~~ | ~~Add `TODO_LIST.md`~~ done at `474f9dd`                        | ~~Low~~ | ~~15 min~~ | ~~Documentation~~ |
| ~~21~~ | ~~Add shell completions~~ done at `81d8c44`                     | ~~Low~~ | ~~20 min~~ | ~~UX~~ |
| ~~22~~ | ~~Add `.npmrc` parsing for registry config~~ moved to TODO_LIST.md #12 | ~~Medium~~ | ~~30 min~~ | ~~Feature parity~~ |
| ~~23~~ | ~~Add coverage threshold to CI (fail if <80%)~~ moved to TODO_LIST.md #18 | ~~Low~~ | ~~5 min~~ | ~~DevOps~~ |
| ~~24~~ | ~~Add dependabot/renovate config~~ done — `.github/dependabot.yml` | ~~Low~~ | ~~10 min~~ | ~~DevOps~~ |
| ~~25~~ | ~~Add performance benchmark to CI~~ moved to ROADMAP.md theme 4 | ~~Low~~ | ~~15 min~~ | ~~DevOps~~ |

## g) Top Question I Cannot Figure Out Myself

**Should we add a `--json` output mode and/or `--registry` flag, or keep this
strictly as a faithful 1:1 port?** → ~~answered by history~~: both shipped at
`e64d3a7` (`--json` as D34, `--registry` as D29); the project chose "improve
the tool" over strict 1:1 parity.

The original `upd` has neither. Adding them would make the tool more useful in
CI pipelines and for non-NPM registries (Verdaccio, Nexus, GitHub Packages).
But it expands scope beyond "port this to Go" into "improve this tool." This is
a product direction decision.

---

## Metrics

| Metric          | Previous Report      | This Report          |
| --------------- | -------------------- | -------------------- |
| Source files    | 9                    | 9 (+6 test files)    |
| Source lines    | ~900                 | ~1,196 (source only) |
| Test files      | 3                    | 6                    |
| Test assertions | 22                   | 67                   |
| Test coverage   | 0% (untested engine) | 86.6%                |
| Dependencies    | 3                    | 3 (unchanged)        |
| Binary size     | 6.6 MB               | 6.6 MB               |
| DevOps          | None                 | flake.nix + CI       |

## File Inventory

```
cmd/upd/main.go              # Entry point (84 lines)
internal/config.go           # CLI flags, help, version (139 lines)
internal/config_test.go      # Flag parsing tests (157 lines)
internal/diff.go             # LCS character diff (82 lines)
internal/diff_test.go        # Diff tests (136 lines)
internal/doc.go              # Package doc (2 lines)
internal/engine.go           # Concurrent fetch + state machine (173 lines)
internal/engine_test.go      # Integration tests with mock registry (267 lines)
internal/errors.go           # Sentinel errors (10 lines)
internal/manifest.go         # Dep extraction, glob, version regex (160 lines)
internal/manifest_test.go    # Manifest, regex, glob tests (170 lines)
internal/npm.go              # RegistryClient, version resolution (119 lines)
internal/packagejson.go      # Formatting-preserving JSON reader/writer (120 lines)
internal/packagejson_test.go # UpdateDependency, GetUpdArgs tests (113 lines)
internal/progress.go         # Progress bar (50 lines)
internal/render.go           # Colored table renderer (234 lines)
internal/render_test.go      # Render output tests (180 lines)
```
