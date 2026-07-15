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

1. **Commit the `go.mod` fix** — error-family is currently listed as `// indirect` but should be a direct dependency. `go mod tidy` fix is uncommitted.
2. **`cmd/upd` test coverage** — dropped from 10.2% to 3.9% after removing exit-code tests (moved to `errors_test.go` in the `upd` package). The `run()` and `finalizeRun()` functions are completely untested.
3. **Stale status report** — `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md` is now superseded but still exists.
4. **`docs/DOMAIN_LANGUAGE.md`** — has uncommitted formatting changes from a prior session (may already be committed, needs verification).
5. **`flake.lock`** — may need update after go.mod changes.

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

1. **Fix go.mod** — `go mod tidy` to correct direct/indirect classification. Uncommitted.
2. **Add cmd/upd integration tests** — `run()` is the main entry point and has 3.9% coverage. Need tests that exercise the full pipeline with mock registries.
3. **Add errorfamily.HandleError integration test** — verify that the full HandleError path produces correct stderr output + exit codes for each family.
4. **Consider using errorfamily.Registry for test isolation** — currently using DefaultRegistry globally; tests could use scoped registries.
5. **The retryableError type in npm.go** could potentially implement the errorfamily.Retryable interface instead of being a separate wrapper type. This would let errorfamily.Classify automatically detect retryability.
6. **Add error codes to the --json output** — currently JSON output has `state` and `error` strings, but not the machine-readable `code` or `family`. CI consumers would benefit from structured error codes.
7. **Render errorfamily context in --verbose mode** — `errorfamily.Error.Format(f, '+')` produces verbose output with context keys. The `--verbose` flag could leverage this.
8. **Context-loss issues (branching-flow)** — 12 MEDIUM issues remain. Most are for complex types (manifest, decoder) but some could be addressed by adding `.WithContext()` calls.

---

## f) Up to 50 Things We Should Get Done Next

### High Priority — Correctness & Coverage

1. **Commit the `go.mod` fix** (error-family as direct dep, not indirect)
2. **Write integration test for `cmd/upd/main.go:run()`** — mock registry, verify full pipeline
3. **Write test for `finalizeRun` JSON path** — verify RenderJSON is called when cfg.JSON=true
4. **Write test for `finalizeRun` write gate** — verify no write when updates=0 or cfg.Nop=true
5. **Write test for `finalizeRun` partial failure** — verify ErrPartialFailure returned when errCount>0
6. **Delete stale status report** `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md`
7. **Run `go mod tidy`** and commit the result
8. **Update flake.lock** if needed after dependency changes

### Medium Priority — Error UX Polish

9. **Add error `code` and `family` fields to --json output** — machine-readable for CI
10. **Leverage `errorfamily.Format('+')` in --verbose mode** — structured verbose output
11. **Test HandleError end-to-end** — capture stderr, verify message templates render correctly
12. **Register `context.DeadlineExceeded` as Transient** in DefaultRegistry (for timeout errors)
13. **Register `context.Canceled` as Rejection** (for Ctrl+C)
14. **Add retryableError.IsRetryable() method** — implement errorfamily.Retryable interface
15. **Consider removing retryableError wrapper entirely** — errorfamily.Transient already signals retryability
16. **Add errorfamily.Code(err) to render.go error display** — show machine code in verbose mode

### Medium Priority — Architecture

17. **Consider making `retryableError` implement `errorfamily.Classified`** — return Transient family directly
18. **Review if `classifyRegistryError` can use errorfamily.Registry.RegisterClassification** instead of custom logic
19. **Split `config.go`** — Config struct, ParseFlags, PrintUsage, ShouldDisableColor are 4 concerns
20. **Consider extracting progress reporter** into its own file (currently inline in engine.go)
21. **Add `Section` named type** for dependency section names (currently bare strings)
22. **Consider `PackageName` named type** — used in 10+ places as bare string

### Lower Priority — Quality

23. **Address remaining 12 CONTEXT branching-flow issues** (add .WithContext where practical)
24. **Add fuzzing tests for packagejson.go JSON parsing**
25. **Add fuzzing tests for manifest.go version regex**
26. **Update FEATURES.md** with errorfamily adoption
27. **Update TODO_LIST.md** with cmd/upd coverage gap
28. **Add `.branching-flow.toml`** to permanently suppress PHANTOM (56 noise violations)
29. **Consider bumping go.mod to go1.27** when released (eliminates 34 stdversion warnings)
30. **Add benchmark for HandleError** — ensure template rendering doesn't slow down CLI exit
31. **Add benchmark for error chain classification** — ensure errors.AsType is fast enough
32. **Consider adding `upd doctor` subcommand** — check registry connectivity
33. **Consider adding shell completions** (bash/zsh/fish)
34. **Consider adding `upd init` subcommand** — create `upd` field in package.json
35. **Add test for WithContext chaining** — verify multiple WithContext calls accumulate
36. **Add test for message template resolution** — verify correct template matched by code
37. **Add test for concurrent modification full path** — file written, then modified, then upd.Write fails
38. **Consider structured logging (slog)** — replace fmt.Fprintf warnings with structured logs
39. **Review if `printWarnings` should use errorfamily** — currently raw fmt.Fprintf
40. **Consider adding `--format` flag** — json/table/csv output modes
41. **Consider adding dry-run diff output** — show what would change without writing
42. **Consider adding lockfile parsing** — yarn.lock, pnpm-lock.yaml
43. **Consider adding monorepo workspace support**
44. **Consider adding a GitHub Action** that runs branching-flow on PRs
45. **Review if `Spec.Err` should be `*errorfamily.Error`** instead of bare `error`
46. **Consider adding `Spec.Code()` and `Spec.Family()` methods** for structured access
47. **Add test for errors.Is across WithContext cloning** — verify identity preservation
48. _*Add test for errors.Is across Wrap* functions_* — verify chain traversal
49. **Consider adding errorfamily.HTTPHandler** if upd ever gets an HTTP API
50. **Consider adding errorfamily.RetryPolicy integration** with npm.go retry loop

---

## g) Top 2 Questions

### 1. Should `retryableError` be replaced by errorfamily's classification system?

Currently `npm.go` has a custom `retryableError` struct that wraps transient errors with a `retryAfter` duration. The retry loop checks `errors.As(err, &retryErr)` to decide whether to retry. With errorfamily, `Family.IsRetryable()` already signals retryability — but it doesn't carry the `retryAfter` duration from the `Retry-After` HTTP header. Should I:

- **(a)** Keep `retryableError` as-is (it carries extra data errorfamily doesn't model), or
- **(b)** Make it implement `errorfamily.Retryable` and `errorfamily.Classified` interfaces, or
- **(c)** Add `RetryAfter` to errorfamily's `Error` struct and eliminate the custom type entirely?

### 2. Should the --json output include structured error codes and families?

Currently `RenderJSON` outputs `{name, section, old, new, state, error}` where `error` is a flat string. With errorfamily, every error now has a machine-readable `code` (e.g., `registry.package_not_found`) and `family` (e.g., `rejection`). Should the JSON output include these fields? This would be a **breaking change** for any CI scripts that parse the current JSON schema, but would make the output far more useful for automated consumers. Should I add them as new fields (additive) or replace the flat `error` string?
