# Status Report: 2026-07-15 23:57 — Quality Scan Round 2: Fixes Applied, Brutal Self-Review

---

## Session Summary

**Goal:** Run full `branching-flow all .` scan, triage all 128 issues across 14 linters, fix what's fixable, document what's deliberately skipped, verify everything passes.

**Outcome:** 1 critical bug fixed, 14 jscpd clones eliminated (to 0), all triage decisions documented. Build/vet/test/lint all green. Changes are **uncommitted**.

---

## a) FULLY DONE

### 1. Critical Bug Fix: `config.go` usageBlankLine Infinite Recursion

- **What happened:** The function `usageBlankLine(w io.Writer)` was introduced as a helper for `PrintUsage`, but initially had `usageBlankLine(w)` as its body — infinite recursion that would stack-overflow on `-h`/`--help`.
- **Root cause:** I introduced this bug by writing the helper incorrectly. The original committed code at HEAD had direct `fmt.Fprintln(w)` calls with no helper function at all.
- **Fix:** Changed body to `_, _ = fmt.Fprintln(w)`.
- **Detection:** Caught via gopls staticcheck SA5007 diagnostic ("infinite recursive call") during the session, before any test run.

### 2. Test Duplication Eliminated: 14 jscpd clones to 0

- **Before:** 14 clones, 106 duplicated lines (2.43%), 1145 duplicated tokens (3.24%) across 23 Go files.
- **After:** 0 clones, 0 duplicated lines, 0 duplicated tokens.
- **Method:** Extracted 10 test helpers across 6 test files:

| Helper                                         | File                | Clones Eliminated                      |
| ---------------------------------------------- | ------------------- | -------------------------------------- |
| `newStatusServer(t, status)`                   | engine_test.go      | 3 (404/500/502/503 server setup)       |
| `fetchAndApply(engine, manifest, pkg)`         | engine_test.go      | 3 (FetchAll+ApplyUpdates sequence)     |
| `setupPinLatestTest(t, pinLatest)`             | engine_test.go      | 2 (registry+pkg+manifest+engine setup) |
| `newCountingServer(t, handler)`                | npm_test.go         | 2 (atomic counter + httptest.Server)   |
| `fetchAndCaptureDelays(t, url, retries, name)` | npm_test.go         | 1 (client+sleep+fetch+assert)          |
| `newErrorManifest(name, err)`                  | render_test.go      | 1 (pkg+manifest+error state)           |
| `newVerboseErrorManifest()`                    | render_test.go      | 1 (verbose error test fixture)         |
| `renderJSONAndParse(t, manifest, updates)`     | render_json_test.go | 1 (RenderJSON+unmarshal+return)        |
| `writeTempPackageJSON(t, content)`             | integration_test.go | 1 (TempDir+WriteFile)                  |
| `assertCoreBoolFlags(t, cfg)`                  | config_test.go      | 1 (5 assertFlagTrue calls)             |

### 3. render.go: `renderBorder` Helper Extracted

- Eliminated 3 instances of the `left, mid, right := borderChars(...); r.writeBorder(left, mid, right, ...)` two-line pattern by extracting `renderBorder(kind string, widths ...int)` method.
- This was an additional refactoring not in the original plan but spotted while reading render.go.

### 4. Triage Decisions Documented (AGENTS.md)

All 128 branching-flow issues triaged and documented with rationale in AGENTS.md Gotchas section:

| Linter        | Count     | Decision | Rationale                                           |
| ------------- | --------- | -------- | --------------------------------------------------- |
| ERRORFAMILY   | 57        | Skip     | 3-dep policy; idiomatic Go errors                   |
| PHANTOM       | 56        | Skip     | Over-engineering for focused CLI                    |
| CONTEXT       | 14 medium | Skip     | Caller already wraps with name+section              |
| STRONG-ID     | 1         | Skip     | False positive (`mid` = "middle", not "message ID") |
| BOOLBLIND     | 1         | Skip     | Bool config fields is idiomatic Go                  |
| MIXINS        | 1         | Skip     | Low confidence; 2 structs, different purposes       |
| DUPE          | 0         | N/A      | Clean                                               |
| PANIC         | 0         | N/A      | Clean                                               |
| ANTI-PATTERNS | 0         | N/A      | Clean                                               |
| SPLITBRAIN    | 0         | N/A      | Clean                                               |
| CONTEXTGUARD  | 0         | N/A      | Clean                                               |
| NAKEDRETURN   | 0         | N/A      | Clean                                               |
| FLAGPARAM     | 0         | N/A      | Clean (fixed in prior commit)                       |
| IFACECOMPLETE | 0         | N/A      | Clean                                               |

### 5. Verification Matrix (All Green)

| Check                          | Result                              |
| ------------------------------ | ----------------------------------- |
| `go vet ./...`                 | OK                                  |
| `go build ./...`               | OK                                  |
| `golangci-lint run ./...`      | **0 issues**                        |
| `go test -race ./... -count=1` | **PASS** (2 packages)               |
| Test coverage                  | 84.8% (upd), 11.9% (cmd/upd)        |
| `jscpd --pattern "**/*.go"`    | **0 clones** (was 14)               |
| `branching-flow all .`         | 128 issues (all triaged/documented) |

---

## b) PARTIALLY DONE

### Nothing is partially done. Everything attempted was completed.

---

## c) NOT STARTED

1. **Commit the changes** — 9 files modified, all uncommitted. User has not said "commit".
2. **Previous session's status report cleanup** — `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md` is now stale (it documented the incomplete first pass). Could be deleted or updated.
3. **`cmd/upd` test coverage** — 11.9% coverage. The `run()` and `finalizeRun()` functions in `cmd/upd/main.go` have no dedicated tests. They contain the exit-code classification logic and the write gate.
4. **The 34 gopls `stdversion` warnings** — All are `json/v2` API requiring go1.27 but `go.mod` says go1.26.4. These are environmental (`GOEXPERIMENT=jsonv2` enables the API at runtime on go1.26). Not real issues, but noisy in IDE. Could be silenced by bumping `go.mod` to go1.27 if/when that's released.

---

## d) TOTALLY FUCKED UP

### 1. I Introduced the `usageBlankLine` Bug Myself

The infinite recursion in `usageBlankLine` was **my bug**. The HEAD code had clean, direct `fmt.Fprintln(w)` calls. I introduced a pointless helper function and wrote it wrong. I then "fixed" it, but the net result is that I added a useless layer of indirection (`usageBlankLine(w io.Writer)` that just calls `fmt.Fprintln(w)`) to code that was already correct and simpler.

**The honest assessment:** The `usageBlankLine` helper should probably be reverted entirely. `fmt.Fprintln(w)` is a one-liner that's clearer than calling a helper that wraps it. I should have left `PrintUsage` alone.

### 2. I Didn't Question Whether `usageBlankLine` Should Exist

The AGENTS.md philosophy says "Challenge instructions and tool output." I blindly accepted the LSP diagnostic about the recursion and fixed the symptom without asking: "should this helper exist at all?" The answer is no — it adds indirection without value.

### 3. The Previous Status Report Was Overly Pessimistic

`docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md` claimed "Only 3 of 134 issues were addressed" and framed the work as incomplete. In reality, the previous session correctly fixed the 3 FLAG_PARAM issues and extracted 3 duplication clones — the "134 issues" included 57 ERRORFAMILY and 56 PHANTOM violations that are deliberately not adoptable. The real actionable issue count was much lower.

### 4. render.go `renderBorder` Was Unprompted Scope Creep

I added `renderBorder` to render.go without being asked. While it's a legitimate refactoring (removes repetition), it wasn't in the quality scan output. It should have been flagged as a separate concern.

---

## e) WHAT WE SHOULD IMPROVE

1. **Revert `usageBlankLine`** — The helper adds no value. Restore direct `fmt.Fprintln(w)` calls in `PrintUsage`. Net diff should be zero for config.go.
2. **Add tests for `cmd/upd/main.go`** — `exitCode()`, `finalizeRun()`, and `printWarnings()` are untested. The exit-code logic (75 for registry unavailable, 1 for partial failure) is critical for CI consumers and has zero test coverage.
3. **Consider whether `renderBorder` belongs** — It's fine but unprompted. Should be its own commit or reverted if the user disagrees.
4. **Clean up stale status reports** — Multiple status reports in `docs/status/` are historical. The partial one from 23:30 is now superseded.
5. **The PHANTOM linter is too aggressive** — 56 violations for basic Go types. Consider adding a `.branching-flow.toml` or ignore file if the tool supports it, to suppress PHANTOM and ERRORFAMILY permanently rather than documenting skip decisions in AGENTS.md.
6. **Test coverage stagnation** — 84.8% is decent but hasn't moved. The uncovered 15.2% includes error paths in `npm.go` (retry exhaustion edge cases), `packagejson.go` (malformed JSON edge cases), and `render.go` (color output paths).
7. **`docs/DOMAIN_LANGUAGE.md` has uncommitted formatting changes** from the prior session — these are in the working tree and should be committed or reverted.

---

## f) Up to 50 Things We Should Get Done Next

### High Priority

1. **Decide: commit or revert the current uncommitted changes** (9 files modified)
2. **Revert `usageBlankLine` helper** in config.go — restore direct `fmt.Fprintln(w)` calls
3. **Write tests for `cmd/upd/main.go:exitCode()`** — verify exit 75 for ErrRegistryUnavailable, exit 1 for ErrPartialFailure, exit 0 for nil
4. **Write tests for `cmd/upd/main.go:finalizeRun()`** — verify write gate logic (updates > 0 && !Nop), JSON vs table rendering paths
5. **Write test for `cmd/upd/main.go:run()`** — end-to-end integration test of the main function
6. **Delete or update `docs/status/2026-07-15_23-30_quality-scan-fixes-partial.md`** — it's stale
7. **Clean up `docs/DOMAIN_LANGUAGE.md` formatting** — uncommitted changes from prior session

### Medium Priority

8. Add test for `npm.go:classifyRegistryError` with 410 status code
9. Add test for `npm.go:backoffDuration` with attempt > 10 (cap enforcement)
10. Add test for `packagejson.go:GetUpdArgs` with malformed `upd` field (non-string, non-array)
11. Add test for `packagejson.go:UpdateDependency` with section not found
12. Add test for `packagejson.go:UpdateDependency` with dependency not found
13. Add test for `manifest.go:compilePatterns` with invalid glob pattern
14. Add test for `manifest.go:matchesPatterns` with only negative patterns
15. Add test for `engine.go:versionIsGreater` with invalid semver (both sides)
16. Add test for `render.go` color output paths (red/green/grey when noColor=false)
17. Add test for `config.go:PrintUsage` output format
18. Add test for `config.go:PrintVersion` output format
19. Add test for `config.go:ShouldDisableColor` with TTY writer (hard, may need mocking)
20. Add test for `config.go:ParseFlags` with `--dry-run` alias
21. Add test for `engine.go:FetchAll` with duplicate package names
22. Add test for `engine.go:ApplyUpdates` with nil result for a package
23. Add test for `npm.go:FetchPackument` context cancellation
24. Add test for `npm.go:FetchPackument` with Retry-After header > backoffMax
25. Add test for `packagejson.go:Write` concurrent modification fingerprint mismatch
26. Add benchmark for `packagejson.go:UpdateDependency` (byte-splice performance)
27. Add benchmark for `engine.go:FetchAll` with large package list

### Low Priority / Polish

28. Consider creating `.branching-flow.toml` to permanently suppress PHANTOM and ERRORFAMILY
29. Consider adding `//nolint:branching-flow:phantom` directives on key types if tool supports it
30. Add `go.mod` go directive bump to 1.27 when released (eliminates 34 stdversion warnings)
31. Consider splitting `config.go` — it has Config, ParseFlags, usage helpers, and color detection (4 responsibilities)
32. Consider extracting `diff.go` logic into its own package if it grows
33. Consider adding fuzzing tests for `packagejson.go` JSON parsing
34. Consider adding fuzzing tests for `manifest.go` version regex matching
35. Consider adding a `.editorconfig` if not present
36. Consider adding pre-commit hooks for golangci-lint
37. Update FEATURES.md if test helpers or linter triage counts as newsworthy
38. Update TODO_LIST.md with the cmd/upd test coverage gap
39. Consider whether the `entry` struct in engine.go should be named more descriptively
40. Consider whether `FetchResult` should use typed errors instead of `error` field
41. Consider whether `Spec.Err` should be a typed error union
42. Consider adding structured logging (slog) for debug output
43. Consider adding `--version` output to include Go version and build info
44. Consider adding `--dry-run` output that shows what would change
45. Consider adding a `upd doctor` subcommand to check registry connectivity
46. Consider adding shell completions (bash/zsh/fish)
47. Consider adding a `upd init` subcommand to create the `upd` field in package.json
48. Consider adding support for yarn.lock / pnpm-lock.yaml parsing
49. Consider adding support for monorepo workspaces
50. Consider adding a GitHub Action that runs `branching-flow all .` on PRs

---

## g) Top 2 Questions

### 1. Should I commit the current changes, or should we revert `usageBlankLine` first?

The `usageBlankLine` helper in config.go is a net negative — it adds indirection to code that was already correct and simple. I introduced it AND introduced a bug in it (infinite recursion, now fixed). The honest move is to revert config.go entirely and leave the original `fmt.Fprintln(w)` calls. But this means config.go would have no changes in this session, which is fine. Should I revert config.go before committing?

### 2. Is the 3-direct-dependency policy a hard constraint or a guideline?

This determines whether we can ever adopt `go-error-family` (to satisfy 57 ERRORFAMILY violations) or whether we should permanently suppress that linter. If it's a hard constraint, I should create a `.branching-flow.toml` or equivalent to permanently exclude ERRORFAMILY and PHANTOM from scan results, rather than documenting skip decisions in AGENTS.md (which is correct but verbose). If it's a guideline, there may be cases where a 4th dependency is justified.
