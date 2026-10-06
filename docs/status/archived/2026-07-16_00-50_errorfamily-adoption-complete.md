# Status Report: 2026-07-16 00:50 — go-error-family Adoption + HandleError + Message Templates

---

## Session Summary

**Goal:** Adopt `go-error-family` across the entire `upd` codebase, register message templates for all error codes, replace hand-rolled exit-code logic with `HandleError`, and eliminate all test duplication.

**Outcome:** 4 commits pushed. Build/vet/test/lint/jscpd all green. branching-flow dropped from 128 to 72 issues. One uncommitted `go.mod` fix (error-family listed as indirect instead of direct).

---

## a) FULLY DONE

### 1. Test Duplication Eliminated (commit `78d0cbf`)

- 14 jscpd clones to **0** via 10 helper extractions across 7 test files
- Helpers: `newStatusServer`, `fetchAndApply`, `setupPinLatestTest`, `newCountingServer`, `fetchAndCaptureDelays`, `newErrorManifest`, `newVerboseErrorManifest`, `renderJSONAndParse`, `writeTempPackageJSON`, `assertCoreBoolFlags`

### 2. go-error-family Adoption (commit `db891d0`)

- **13 domain sentinels** rewritten from `errors.New` to `errorfamily.NewRejection/NewCorruption/NewTransient/NewConflict`
- **30+ `fmt.Errorf` wrapping sites** across 6 files converted to `errorfamily.Wrap*` with structured `.WithContext(key, value)`
- **Exit codes now context-aware**: Rejection=1, Transient=75 (EX_TEMPFAIL), Corruption=65 (EX_DATAERR), Conflict=1
- `ErrHelp`/`ErrVersion` kept as plain `errors.New` (control-flow signals, not domain errors)
- depguard, wrapcheck, and vendorHash all updated
- Previous "3-dependency policy" was **fabricated** by prior session — never a real constraint

### 3. Message Templates + HandleError (commit `3cd313e`)

- **`messages.go`**: Registered What/Why/Fix/WayOut templates for all 13 error codes
- **`cmd/upd/main.go`**: Replaced hand-rolled `fmt.Fprintf(os.Stderr, "ERROR: %v")` + `exitCode()` with `errorfamily.HandleError(err)` — one call classifies, formats with templates, writes to stderr, returns exit code
- **`errors_test.go`**: 16 test cases covering every sentinel's Family + ExitCode + wrapped-chain preservation
- Cleaned up `finalizeRun` — removed redundant if/else branches

### 4. Documentation (commit `64d174c`)

- AGENTS.md updated: deps list (3 to 4), error classification description, linter triage decisions

### 5. Verification Matrix (All Green)

| Check                          | Result                              |
| ------------------------------ | ----------------------------------- |
| `go vet ./...`                 | OK                                  |
| `go build ./...`               | OK                                  |
| `golangci-lint run ./...`      | **0 issues**                        |
| `go test -race ./... -count=1` | **PASS** (2 packages)               |
| `jscpd`                        | **0 clones** (25 files, 4513 lines) |
| `nix build .#default`          | OK                                  |
| Test coverage (upd)            | 84.8%                               |
| Test coverage (cmd/upd)        | 3.9%                                |
| branching-flow                 | **72 issues** (was 128)             |

---

## b) PARTIALLY DONE

### Nothing partially done — all attempted work was completed and pushed.

---

## c) NOT STARTED

1. ~~**Commit the `go.mod` fix** — error-family is currently listed as `// indirect` but should be a direct dependency. `go mod tidy` fix is uncommitted.~~ done at `571e312`
2. ~~**`cmd/upd` test coverage** — dropped from 10.2% to 3.9% after removing exit-code tests (moved to `errors_test.go` in the `upd` package). The `run()` and `finalizeRun()` functions are completely untested.~~ partially done — CLI tests added at `ea6493d`/`1cd0109`; the `run()` end-to-end gap moved to TODO_LIST.md #14
3. ~~**Stale status report** — `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md` is now superseded but still exists.~~ done (docs-health pass 2026-09-25 — annotated in place; kept as historical snapshot per docs-health rules)
4. ~~**`docs/DOMAIN_LANGUAGE.md`** — has uncommitted formatting changes from a prior session (may already be committed, needs verification).~~ done — committed; rewritten in `32c208b`
5. ~~**`flake.lock`** — may need update after go.mod changes.~~ done — refreshed repeatedly since (latest `f020505`)

---

## d) TOTALLY FUCKED UP

### 1. Introduced usageBlankLine Infinite Recursion Bug

Created a helper function `usageBlankLine(w io.Writer)` that called itself instead of `fmt.Fprintln(w)`. This would have stack-overflowed on `upd -h`. Caught by staticcheck SA5007 during the session, fixed, then reverted entirely (the helper added no value).

### 2. Lost User-Friendly Concurrent Modification Message

When rewriting `finalizeRun`, both branches of an if/else returned the same `err` — the user-friendly "Your file was not changed" message was lost. Fixed by registering a message template with that text instead.

### 3. Broke .golangci.yml YAML Structure

When adding wrapcheck config, accidentally moved the `ignore-names` list from `varnamelen` to `wrapcheck`, causing 29 spurious varnamelen warnings. Fixed by restoring the correct YAML structure.

### 4. Fabricated "3-Dependency Policy"

Prior session invented a fictional policy to justify skipping ERRORFAMILY. This was never a real constraint — I should have evaluated the library on its merits from the start.

### 5. Coverage Regression in cmd/upd

Moving exit-code tests to the `upd` package improved that package's coverage but cratered `cmd/upd` coverage from 10.2% to 3.9%. The `run()` function — the entire CLI entry point — has zero test coverage.

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**Fix go.mod** — `go mod tidy` to correct direct/indirect classification. Uncommitted.~~ done at `571e312`
2. ~~**Add cmd/upd integration tests** — `run()` is the main entry point and has 3.9% coverage. Need tests that exercise the full pipeline with mock registries.~~ moved to TODO_LIST.md #14 — still open
3. ~~**Add errorfamily.HandleError integration test** — verify that the full HandleError path produces correct stderr output + exit codes for each family.~~ moved to TODO_LIST.md #14
4. ~~**Consider using errorfamily.Registry for test isolation** — currently using DefaultRegistry globally; tests could use scoped registries.~~ Won't implement — 14 static sentinels, no registration collisions; scoped registries add ceremony without a problem
5. ~~**The retryableError type in pnpm.go** could potentially implement the errorfamily.Retryable interface instead of being a separate wrapper type. This would let errorfamily.Classify automatically detect retryability.~~ Won't implement — g1 resolved as option (a): the wrapper stays because it carries `retryAfter`, which errorfamily doesn't model (AGENTS.md)
6. ~~**Add error codes to the --json output** — currently JSON output has `state` and `error` strings, but not the machine-readable `code` or `family`. CI consumers would benefit from structured error codes.~~ moved to TODO_LIST.md #13
7. ~~**Render errorfamily context in --verbose mode** — `errorfamily.Error.Format(f, '+')` produces verbose output with context keys. The `--verbose` flag could leverage this.~~ moved to TODO_LIST.md #13
8. ~~**Context-loss issues (branching-flow)** — 12 MEDIUM issues remain. Most are for complex types (manifest, decoder) but some could be addressed by adding `.WithContext()` calls.~~ done — remaining issues intentionally suppressed (AGENTS.md gotcha, `64d174c`)

---

## f) Up to 50 Things We Should Get Done Next

> docs-health 2026-09-25: the actionable subset of this list now lives in
> `TODO_LIST.md` / `ROADMAP.md`. Resolved items are struck below; the rest
> are raw ideas retained for context.

### High Priority — Correctness & Coverage

1. ~~**Commit the `go.mod` fix** (error-family as direct dep, not indirect)~~ done at `571e312`
2. ~~**Write integration test for `cmd/upd/main.go:run()`** — mock registry, verify full pipeline~~ moved to TODO_LIST.md #14
3. ~~**Write test for `finalizeRun` JSON path** — verify RenderJSON is called when cfg.JSON=true~~ moved to TODO_LIST.md #14
4. ~~**Write test for `finalizeRun` write gate** — verify no write when updates=0 or cfg.Nop=true~~ moved to TODO_LIST.md #14
5. ~~**Write test for `finalizeRun` partial failure** — verify ErrPartialFailure returned when errCount>0~~ done at `077f325` (ErrPartialFailure implemented + tested via `errors_test.go`); finalizeRun-direct tests still tracked in TODO_LIST.md #14
6. ~~**Delete stale status report** `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md`~~ done differently (docs-health pass 2026-09-25 — annotated and kept; deletion replaced by annotation)
7. ~~**Run `go mod tidy`** and commit the result~~ done at `571e312`
8. ~~**Update flake.lock** if needed after dependency changes~~ done (`f020505` and successors)

### Medium Priority — Error UX Polish

9. ~~**Add error `code` and `family` fields to --json output** — machine-readable for CI~~ done at `827b163`
10. ~~**Leverage `errorfamily.Format('+')` in --verbose mode** — structured verbose output~~ done at `e64d3a7` (D40) — `%+v` traverses the chain and errorfamily's `Format` renders the context
11. ~~**Test HandleError end-to-end** — capture stderr, verify message templates render correctly~~ NOT-DO — superseded: upd never calls `errorfamily.HandleError`; fang renders errors and `errorfamily.ExitCode` classifies at the process boundary (covered by CLI e2e tests, `8f8d6fd`)
12. ~~**Register `context.DeadlineExceeded` as Transient** in DefaultRegistry (for timeout errors)~~ Won't implement — timeouts already flow through the transient path (AGENTS.md); a Registry entry would add a second classification mechanism beside `classifyRegistryError`
13. ~~**Register `context.Canceled` as Rejection** (for Ctrl+C)~~ done differently — cancellation during backoff is wrapped as Rejection `registry.fetch_aborted` at the abort site (`npm.go:FetchPackument`)
14. ~~**Add retryableError.IsRetryable() method** — implement errorfamily.Retryable interface~~ Won't implement — g1 option (a)
15. ~~**Consider removing retryableError wrapper entirely** — errorfamily.Transient already signals retryability~~ Won't implement — the wrapper carries `retryAfter`, which errorfamily doesn't model (g1 option a)
16. ~~**Add errorfamily.Code(err) to render.go error display** — show machine code in verbose mode~~ Won't implement — verbose shows full `%+v` chains; machine codes live in the JSON output (`827b163`)

### Medium Priority — Architecture

17. ~~**Consider making `retryableError` implement `errorfamily.Classified`** — return Transient family directly~~ Won't implement — g1 option (a)
18. ~~**Review if `classifyRegistryError` can use errorfamily.Registry.RegisterClassification** instead of custom logic~~ Won't implement — a two-line function with explicit statuses is clearer than registry indirection
19. ~~**Split `config.go`** — Config struct, ParseFlags, PrintUsage, ShouldDisableColor are 4 concerns~~ Won't implement — one cohesive CLI-surface concern; usage helpers were removed by the fang migration
20. ~~**Consider extracting progress reporter** into its own file (currently inline in engine.go)~~ done already — `progress.go` exists with the Reporter interface
21. ~~**Add `Section` named type** for dependency section names (currently bare strings)~~ Won't implement — TODO_LIST R3
22. ~~**Consider `PackageName` named type** — used in 10+ places as bare string~~ Won't implement — TODO_LIST R4

### Lower Priority — Quality

23. ~~**Address remaining 12 CONTEXT branching-flow issues** (add .WithContext where practical)~~ done — addressed at creation sites; the rest intentionally suppressed (AGENTS.md, `64d174c`)
24. ~~**Add fuzzing tests for packagejson.go JSON parsing**~~ Won't implement — trusted local input; decoder error paths unit-tested
25. ~~**Add fuzzing tests for manifest.go version regex**~~ Won't implement — property tests cover the regex space (`manifest_property_test.go`)
26. ~~**Update FEATURES.md** with errorfamily adoption~~ done (docs-health pass 2026-09-25)
27. ~~**Update TODO_LIST.md** with cmd/upd coverage gap~~ done (docs-health pass 2026-09-25 — TODO_LIST #14)
28. ~~**Add `.branching-flow.toml`** to permanently suppress PHANTOM (56 noise violations)~~ Won't implement — triage rationale documented in AGENTS.md instead
29. ~~**Consider bumping go.mod to go1.27** when released (eliminates 34 stdversion warnings)~~ done at `d43f460`
30. ~~**Add benchmark for HandleError** — ensure template rendering doesn't slow down CLI exit~~ Won't implement — exit-path rendering runs once per process, and upd doesn't call HandleError (see f11)
31. ~~**Add benchmark for error chain classification** — ensure errors.AsType is fast enough~~ Won't implement — classification over a 2-level chain is nanoseconds; no hot path
32. ~~**Consider adding `upd doctor` subcommand** — check registry connectivity~~ moved to ROADMAP.md theme 2
33. ~~**Consider adding shell completions** (bash/zsh/fish)~~ done at `81d8c44` (Cobra `completion` command)
34. ~~**Consider adding `upd init` subcommand** — create `upd` field in package.json~~ Won't implement — one JSON line; a command would be ceremony
35. ~~**Add test for WithContext chaining** — verify multiple WithContext calls accumulate~~ NOT-DO — library-owned surface; that contract belongs to go-error-family's own test suite
36. ~~**Add test for message template resolution** — verify correct template matched by code~~ NOT-DO — library-owned surface (same rationale)
37. ~~**Add test for concurrent modification full path** — file written, then modified, then upd.Write fails~~ done — `TestWriteRejectsConcurrentModification` (`packagejson_test.go:277`) plus the `messages.go` recovery template
38. ~~**Consider structured logging (slog)** — replace fmt.Fprintf warnings with structured logs~~ Won't implement — TODO_LIST R12
39. ~~**Review if `printWarnings` should use errorfamily** — currently raw fmt.Fprintf~~ Won't implement — warnings are display lines, not classified errors; the yellow `WARNING:` format is the tested contract
40. ~~**Consider adding `--format` flag** — json/table/csv output modes~~ done at `2ea7868` — `--format=table|json` (csv rejected: JSON covers machine consumption)
41. ~~**Consider adding dry-run diff output** — show what would change without writing~~ already in ROADMAP theme 2 (dry-run renders the old→new table today)
42. ~~**Consider adding lockfile parsing** — yarn.lock, pnpm-lock.yaml~~ already in ROADMAP theme 3
43. ~~**Consider adding monorepo workspace support**~~ already in ROADMAP theme 3
44. ~~**Consider adding a GitHub Action** that runs branching-flow on PRs~~ Won't implement — local audit tool, not a CI-grade dependency
45. ~~**Review if `Spec.Err` should be `*errorfamily.Error`** instead of bare `error`~~ Won't implement — the field holds wrapped chains (errors.Join at `3696a33`); the interface type is correct
46. ~~**Consider adding `Spec.Code()` and `Spec.Family()` methods** for structured access~~ Won't implement — the JSON renderer extracts code/family where consumed (`render.go`); accessor methods without callers are ceremony
47. ~~**Add test for errors.Is across WithContext cloning** — verify identity preservation~~ NOT-DO — library-owned surface (go-error-family test suite)
48. ~~**Add test for errors.Is across Wrap_ functions** — verify chain traversal~~ NOT-DO — library-owned surface (same rationale)
49. ~~**Consider adding errorfamily.HTTPHandler** if upd ever gets an HTTP API~~ Won't implement — no HTTP surface; ROADMAP non-goals exclude a server mode
50. ~~**Consider adding errorfamily.RetryPolicy integration** with pnpm.go retry loop~~ Won't implement — the custom loop honors `Retry-After`, which a generic policy doesn't express (g1 option a)

---

## g) Top 2 Questions

### 1. Should `retryableError` be replaced by errorfamily's classification system?

Currently `pnpm.go` has a custom `retryableError` struct that wraps transient errors with a `retryAfter` duration. The retry loop checks `errors.As(err, &retryErr)` to decide whether to retry. With errorfamily, `Family.IsRetryable()` already signals retryability — but it doesn't carry the `retryAfter` duration from the `Retry-After` HTTP header. Should I:

- **(a)** Keep `retryableError` as-is (it carries extra data errorfamily doesn't model), or
- **(b)** Make it implement `errorfamily.Retryable` and `errorfamily.Classified` interfaces, or
- **(c)** Add `RetryAfter` to errorfamily's `Error` struct and eliminate the custom type entirely?

> **Resolved:** option (a) — the wrapper stays because it carries `retryAfter` from the `Retry-After` header, which errorfamily doesn't model. Documented in AGENTS.md.

### 2. Should the --json output include structured error codes and families?

Currently `RenderJSON` outputs `{name, section, old, new, state, error}` where `error` is a flat string. With errorfamily, every error now has a machine-readable `code` (e.g., `registry.package_not_found`) and `family` (e.g., `rejection`). Should the JSON output include these fields? This would be a **breaking change** for any CI scripts that parse the current JSON schema, but would make the output far more useful for automated consumers. Should I add them as new fields (additive) or replace the flat `error` string?

> **Resolved:** done at `827b163` — error entries carry additive `code` and `family` fields alongside the human `error` string.
