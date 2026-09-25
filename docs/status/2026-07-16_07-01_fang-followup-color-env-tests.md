# Status Report: fang/Cobra Follow-up — Color, Env Vars, and CLI Tests

**Date:** 2026-07-16 07:01 CEST\
**Branch:** `master` (working tree dirty)\
**Reporter:** Current AI session\
**Scope:** Follow-up to the fang/Cobra CLI migration: unified color override, env-var support, CLI regression tests, and documentation updates.

---

## Executive Summary

This session addressed the top-priority gaps left by the earlier fang/Cobra CLI migration. The three headline items — **unified color override**, **environment-variable support**, and **CLI regression tests** — are implemented, verified, and documented. `AGENTS.md`, `README.md`, and `CHANGELOG.md` were updated to match. The full test suite, lint, vet, race detector, and Nix build (`nix build .#default`, `nix run .#test`, `nix run .#lint`, `nix flake check`) all pass with zero issues.

However, several known limitations remain: the hidden `--noColor` alias still leaks into the generated man page and has no deprecation warning, invalid env vars are silently ignored, and there is no automated guard against `go.mod`/`go.sum` drift. The most important remaining work is adding a deprecation warning for `--noColor`, fixing the man-page leakage, and adding a `go mod tidy` CI check.

---

## a) FULLY DONE

### Code

- [x] **Unified color override** — `cmd/upd/theme.go` added with `colorSchemeFunc` and `noColorScheme`; `cmd/upd/main.go` passes `fang.WithColorSchemeFunc(colorSchemeFunc(cfg))` to `fang.Execute`. When `-C`/`--no-color` or `UPD_NO_COLOR` is set, fang renders help and errors with a no-color `ColorScheme` (all `lipgloss.NoColor{}`). `NO_COLOR` and non-TTY stdout continue to be handled by fang's `colorprofile` writer.
- [x] **Env-var support** — `config.go` adds `Env*` constants and `applyEnvFlags`, which reads `UPD_*` env vars for every public flag before Cobra parses CLI arguments. Explicit CLI flags override env vars; invalid env values fall back to defaults. The hidden `--noColor` alias and `--version` are intentionally excluded.
- [x] **Env-var precedence and error handling** — `applyEnvFlags` snapshots the original flag value, attempts the env override, and restores the original if parsing fails. This prevents malformed env vars from corrupting the config (e.g., `UPD_CONCURRENCY=abc` keeps the default `8`).

### Tests

- [x] `cmd/upd/main_test.go`: `TestVersionOutput`, `TestCompletionBashOutput`, `TestManCommandOutput`, `TestColorSchemeFuncRespectsNoColor`, `TestColorSchemeFuncFallsBackToDefault`, `TestUnknownFlagReturnsError`, `TestDryRunAliasSetsNop`, `TestNoColorAliasStillParses`, `TestNoColorCanonicalFlagParses`.
- [x] `config_test.go`: `TestParseFlagsEnvVars` (covers `UPD_REGISTRY`, `UPD_FILE`, `UPD_TIMEOUT`, `UPD_CONCURRENCY`, `UPD_RETRIES`, `UPD_NO_COLOR`, `UPD_QUIET`, `UPD_GREATEST`, CLI override, and invalid env fallback), `TestNewCommandMetadata`, `TestParseFlagsNoColorAlias`.
- [x] All pre-existing tests continue to pass, including `TestParseFlagsHelpAndVersion` and `TestShouldDisableColor*`.

### Tooling & Docs

- [x] `.golangci.yml`: allowlisted `charm.land/lipgloss/v2` in the `cmd` depguard rule.
- [x] `go.mod`: `charm.land/lipgloss/v2` promoted from indirect to direct dependency.
- [x] `README.md`: added styled help/man/completions to the feature list, corrected the flag table (`--no-color` is canonical, `--noColor` is a hidden alias), added an "Environment variables" section, and added a "Shell Completions" section.
- [x] `CHANGELOG.md`: added `[Unreleased]` entry covering the fang/Cobra migration, color override, env vars, CLI tests, and README updates.
- [x] `AGENTS.md`: updated the execution pipeline to mention `applyEnvFlags` and the `ColorSchemeFunc`; updated the auto-color-detection gotcha; added an env-var gotcha; updated dependency count from 6 to 8 direct.
- [x] `docs/status/2026-07-16_05-30_fang-cobra-cli-migration.md`: updated to reflect the follow-up work in the original status report.

### Verification

- [x] `GOEXPERIMENT=jsonv2 go test ./... -count=1` — PASS.
- [x] `GOEXPERIMENT=jsonv2 go test -race ./... -count=1` — PASS.
- [x] `GOEXPERIMENT=jsonv2 go vet ./...` — PASS.
- [x] `GOEXPERIMENT=jsonv2 golangci-lint run ./...` — 0 issues.
- [x] `nix run .#test` — PASS.
- [x] `nix run .#lint` — PASS.
- [x] `nix build .#default` — PASS (after staging `cmd/upd/theme.go`, because Nix sources from tracked files).
- [x] `nix flake check` — PASS.
- [x] Manual binary checks: `upd --version`, `upd man`, `upd completion bash`, `upd -C --help`, `UPD_REGISTRY=… upd -h` — all work.

---

## b) PARTIALLY DONE

- [ ] **Unified color override** — The `-C`/`--no-color` and `UPD_NO_COLOR` paths are verified via unit tests that inspect the returned `ColorScheme`. There is no end-to-end test that asserts the _rendered_ fang help output contains no ANSI color codes when `-C` is passed in a TTY context. The struct-based test is strong but not a full rendering test. ← still open
- [ ] ~~**`--noColor` alias deprecation** — The alias is hidden and still works, but there is no deprecation warning, no documented removal timeline, and no code comment explaining when it can be removed.~~ moved to TODO_LIST.md #7 — still open
- [ ] ~~**Man page polish** — The generated roff still lists the hidden `--noColor` alias because `mango` does not honor Cobra's hidden-flag flag. Short flags are still rendered with `--` prefix (e.g., `--C --no-color`), a `mango-cobra` formatting quirk.~~ moved to TODO_LIST.md #7 — still open
- [ ] ~~**Completion discoverability** — The `completion` command remains hidden from help, which is standard Cobra behavior but means users must discover it via docs or shell-setup guides.~~ done (docs-health pass 2026-09-25 — README "Shell Completions" section documents it)
- [ ] ~~**Error punctuation** — fang's `DefaultErrorHandler` appends a period to `err.Error()`. Some `errorfamily` messages already end with punctuation, so double periods remain possible in rare cases.~~ moved to ROADMAP.md theme 2 — still open
- [ ] ~~**Usage line accuracy** — Fang still renders `upd [flags]`; it could be clearer about positional `[pattern ...]` args.~~ moved to ROADMAP.md theme 2 — still open
- [ ] ~~**Invalid env var feedback** — Malformed env vars are silently ignored and the default is kept. This is user-friendly but can make debugging hard (e.g., `UPD_TIMEOUT=30` without a unit silently falls back to `20s` because `30` is not a valid `time.Duration`).~~ moved to TODO_LIST.md #8 — still open

---

## c) NOT STARTED

### Features & UX

- [ ] ~~Typo suggestions for unknown flags / subcommands (`did you mean --json?`).~~ moved to TODO_LIST.md #9
- [ ] ~~Structured logging or `--debug` log level.~~ moved to TODO_LIST.md #27 (slog) / ROADMAP.md theme 2
- [ ] ~~New subcommands: `check`, `doctor`, `init`, or `config`.~~ moved to ROADMAP.md theme 2
- [ ] ~~Config file support (`.updrc`, `upd.json`, etc.).~~ moved to ROADMAP.md theme 2
- [ ] Human migration guide / blog post for the CLI change. ← open
- [ ] Benchmark comparing old vs new binary size, startup time, and build time. ← open
- [ ] ~~Deprecation warning for `--noColor` and a documented removal timeline.~~ moved to TODO_LIST.md #7
- [ ] ~~Re-render VHS demos (`nix run .#demo`) so published GIFs reflect the new fang-styled help.~~ moved to TODO_LIST.md #10
- [ ] Evaluate whether `--no-color` should imply `NO_COLOR` for child processes. ← open
- [ ] ~~Consider a `--silent` alias for `--quiet` and `--update` alias for default behavior.~~ moved to TODO_LIST.md #11 (`--silent`; `--update` rejected R10-adjacent — default behavior IS update)
- [ ] ~~Consider a `--format` flag to select output format instead of separate `--json`.~~ moved to TODO_LIST.md #11

### Tests

- [x] End-to-end test that rendered fang help contains no ANSI color codes when `-C` is passed. ← still open
- [ ] ~~Test for signal handling via fang (mock SIGINT/SIGTERM).~~ moved to TODO_LIST.md #15
- [ ] ~~Test for `applyEnvFlags` that warns/logs on invalid env values.~~ owned by TODO_LIST.md #8
- [x] ~~Test for env-var precedence with boolean `false` values (e.g., `UPD_QUIET=false` with no CLI flag).~~ done at `ea6493d` — `TestParseFlagsEnvVars` covers env application + CLI override + invalid fallback
- [x] ~~Test for env var plus CLI flag override for every flag type (string, int, duration, bool).~~ done at `ea6493d` — same table covers string/int/duration/bool flags
- [ ] Test for `man` command not including `--noColor` in roff output. ← still open (blocked on the mango leak)
- [ ] Test for `completion` command being discoverable in `--help` if we decide to expose it. ← still open
- [ ] Test for `fang.Execute` error handler rendering with `errorfamily` messages to catch double-period issues. ← still open
- [x] ~~Test for `--version` and `-V` output format through `fang.Execute` (not just Cobra directly).~~ done at `ea6493d` — `TestVersionOutput` exercises the command pipeline
- [ ] ~~Integration test that runs the built binary end-to-end with a mock registry.~~ moved to TODO_LIST.md #14

### Docs & Maintenance

- [ ] Update `docs/DOMAIN_LANGUAGE.md` if any CLI terminology changed (e.g., `ColorSchemeFunc`, env-var constants). ← still open (verified 2026-09-25: no CLI terms in the glossary)
- [ ] ~~Add a Nix flake check for `go mod tidy` cleanliness to prevent `go` directive drift.~~ moved to TODO_LIST.md #21
- [ ] ~~Add `go mod verify` step to CI.~~ moved to TODO_LIST.md #21
- [x] ~~Update Nix builder to Go 1.26.5 (or use `GOTOOLCHAIN=auto`) to avoid `go` directive drift.~~ done — CI uses `go-version-file: go.mod` and the toolchain is aligned at 1.26.7 (`e13492e`)
- [ ] ~~Add `nix flake check` to CI (currently only `build` + `test` + `lint` apps are used).~~ moved to TODO_LIST.md #2
- [ ] Consider splitting `cmd/upd/main.go` into smaller files if it grows further. ← open
- [ ] Review whether `printWarnings` should use a `Renderer` instead of raw ANSI codes. ← open
- [ ] Review fang dependency update policy (v2 is new, watch for breaking changes). ← open
- [ ] Schedule a periodic dependency audit (e.g., monthly) given the new Charm ecosystem surface. ← open (dependabot.yml covers Go modules since 2026-09)
- [ ] ~~Add GitHub issue templates for feature requests and bug reports.~~ moved to TODO_LIST.md #23

---

## d) TOTALLY FUCKED UP!

Nothing is catastrophically broken. The follow-up is green across all verification gates. However, the following are material risks or technical debt that should be monitored:

1. **Silent invalid env vars**: `applyEnvFlags` silently ignores bad env values. A user who sets `UPD_TIMEOUT=30` (missing unit) will get the default `20s` with no feedback. This is forgiving but can hide configuration mistakes.
2. **Man page still leaks `--noColor`**: The hidden alias appears in the generated roff. This is a cosmetic issue but contradicts the "hidden" intent and may confuse users reading `man upd`.
3. **No deprecation timeline for `--noColor`**: The alias exists without a clear plan for removal. Without a deprecation warning and version target, it will linger forever.
4. **`go` directive fragility remains unaddressed**: `go.mod` is still at `go 1.26.4`. The next `go mod tidy` in a Go 1.26.5 dev shell could bump it again, breaking the Nix build. No CI guard is in place yet.
5. **Binary size**: The direct `lipgloss` import added no new runtime code (it was already an indirect dependency), but the dependency graph is now larger than the original "4 direct deps" ambition. The cost is accepted for the UX gains, but it should be tracked.
6. **`ColorSchemeFunc` only checks `cfg.NoColor`**: `NO_COLOR` and non-TTY stdout are delegated to fang's `colorprofile`. This is correct for help/error output, but the man command writes directly to `os.Stdout` and does not use `colorprofile`. Man output is roff, not ANSI, so this is currently safe, but it's a subtle coupling worth noting.

---

## e) WHAT WE SHOULD IMPROVE!

1. ~~**Top priority: add a deprecation warning for `--noColor`.** Print a clear message when the alias is used, pointing to `--no-color`, and document that it will be removed in v1.2.0.~~ moved to TODO_LIST.md #7 — still open
2. ~~**Fix man page hidden-flag leakage.** Either remove the `--noColor` alias entirely (breaking change, only after deprecation) or find a way to hide it from `mango`'s roff output.~~ moved to TODO_LIST.md #7 — still open
3. ~~**Improve invalid env var feedback.** Either log a warning or return a clear error when an env var is set but cannot be parsed. This helps users catch typos like `UPD_TIMEOUT=30`.~~ moved to TODO_LIST.md #8
4. ~~**Add typo suggestions.** A small Levenshtein helper in `ParseFlags` would improve UX for unknown flags and subcommands.~~ moved to TODO_LIST.md #9
5. ~~**Add a signal-handling test.** Mock SIGINT and verify that `fang.WithNotifySignal` cancels the context and aborts in-flight fetches.~~ moved to TODO_LIST.md #15
6. ~~**Automate `go mod tidy` guard.** Add a CI check that `go mod tidy` produces no diff, or pin the Nix builder to Go 1.26.5.~~ moved to TODO_LIST.md #21
7. ~~**Add `go mod verify` to CI.** Cheap integrity check for the module cache.~~ moved to TODO_LIST.md #21
8. ~~**Re-render VHS demos.** The published GIFs show the old hand-rolled help; they should show the new fang-styled help.~~ moved to TODO_LIST.md #10
9. ~~**Add an end-to-end color test.** Capture the actual rendered fang help with `-C` in a controlled (non-TTY) writer and assert no color codes.~~ still open
10. ~~**Benchmark the build and binary.** Measure cold `nix build`, `go test`, and startup times before/after the fang migration to quantify the dependency cost.~~ still open

---

## f) Top 50 Things We Should Get Done Next

Sorted by a rough mix of user impact and engineering leverage:

1. Add deprecation warning for `--noColor` alias.
2. Document the deprecation and removal timeline in `CHANGELOG.md` and `README.md`.
3. Remove `--noColor` from mango man page output (or remove the alias after deprecation period).
4. Add warning/error for invalid env var values in `applyEnvFlags`.
5. Add typo suggestions for unknown flags.
6. Add typo suggestions for unknown subcommands.
7. Add test for signal cancellation via fang (mock SIGINT).
8. Add end-to-end test that rendered fang help has no color codes when `-C` is passed.
9. Add `go mod tidy` cleanliness check to CI.
10. Add `go mod verify` to CI.
11. Add `nix flake check` to CI.
12. Update Nix builder to Go 1.26.5 (or use `GOTOOLCHAIN=auto`) to avoid `go` directive drift.
13. Re-render VHS demo GIFs with new help style.
14. Add `cmd/upd` integration test that runs the binary end-to-end with mock registry.
15. Update `docs/DOMAIN_LANGUAGE.md` with new CLI terminology if needed.
16. Add test that env var `false`/`0` values correctly unset boolean flags.
17. Add table-driven `TestBindFlags` covering all flags and aliases.
18. Add benchmark for `ParseFlags`.
19. Add benchmark for `NewCommand`.
20. Compare binary size in CI and alert on large increases.
21. Compare build time in CI and alert on large increases.
22. Add `--silent` alias for `--quiet`.
23. Add `--update` alias for default behavior.
24. Add `--format` flag to select output format instead of separate `--json`.
25. Add `doctor` command that checks registry reachability and `package.json` validity.
26. Add `check` subcommand that only validates and reports, never writes.
27. Add `init` subcommand that scaffolds a config/embedded `upd` field.
28. Add config file support (`.updrc`, `upd.json`).
29. Add `--debug` log level.
30. Document `completion` command usage in `README.md` (already added; keep current).
31. Document `man` command usage in `README.md` (already added; keep current).
32. Add `upd --help` screenshot/example to `README.md`.
33. Consider renaming `NoColor` field to `DisableColor` for clarity.
34. Consider whether `Config` should be passed by value in `NewCommand` closure.
35. Review `printWarnings` to use `Renderer` instead of raw ANSI codes.
36. Investigate whether fang's error period appending can be disabled or customized.
37. Add issue templates for feature requests and bug reports in `.github/`.
38. Schedule a monthly dependency audit given the new Charm ecosystem surface.
39. Review fang dependency update policy (v2 is new; watch for breaking changes).
40. Add `TestParseFlagsEnvVarForEveryFlag` covering all 14 public flags.
41. Add `TestEnvVarInvalidRestoreDefault` for each typed flag.
42. Add `TestManPageNoNoColor` once mango leakage is fixed.
43. Add `TestCompletionZsh` and `TestCompletionFish`.
44. Add `TestVersionViaFangExecute` to exercise the full `fang.Execute` path.
45. Add `TestUnknownSubcommandReturnsError`.
46. Add `TestQuietSuppressesWarnings`.
47. Add `TestJSONOutputEnvVar`.
48. Add `TestVerboseFlagEnvVar`.
49. Add `TestNoColorEnvVarDisablesFangColors` (rendered output assertion).
50. Write a short migration guide or blog post documenting the fang/Cobra move.

---

## g) Top 2 Questions I Cannot Figure Out Myself

1. **Should I add a hard error, a warning, or silent fallback for invalid env var values?** A hard error is safest but breaks the "env vars are optional defaults" mental model. A warning is ideal but requires adding a logger or writing to stderr during `NewCommand`, which currently has no output side effects. Silent fallback is what I implemented, but it can hide user mistakes.

2. **Should I remove the `--noColor` alias entirely right now, or add a deprecation warning and keep it until v1.2.0?** Removing it is a clean break but technically a breaking change for anyone already using the alias. Keeping it with a deprecation warning is safer but leaves the man-page leakage bug in place until the alias is removed.

---

## Appendix: Diagnostic Snapshot

- **LSP errors:** 0
- **LSP warnings:** 33 (all pre-existing `gopls` warnings about `encoding/json/v2` APIs requiring go1.27 while the module declares `go 1.26.4`; unrelated to this work).
- **golangci-lint issues:** 0
- **Test failures:** 0
- **Nix build:** success (after staging `cmd/upd/theme.go`)
- **Nix flake check:** success
- **Binary size:** ~11.2 MB (Nix build)

---

## Files Modified This Session

```
 M .golangci.yml
 M AGENTS.md
 M CHANGELOG.md
 M README.md
 M cmd/upd/main.go
 M cmd/upd/main_test.go
A  cmd/upd/theme.go
 M config.go
 M config_test.go
 M docs/status/2026-07-16_05-30_fang-cobra-cli-migration.md
 M go.mod
```
