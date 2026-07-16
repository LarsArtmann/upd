# Status Report: fang/Cobra CLI Migration

**Date:** 2026-07-16 05:30 CEST  
**Branch:** `master` (working tree dirty)  
**Reporter:** Current AI session  
**Scope:** Migration of `upd` CLI layer from stdlib `flag` to `charm.land/fang/v2` + Cobra.

---

## Executive Summary

This session migrated `upd` from a hand-rolled stdlib `flag` CLI to a Cobra-based CLI styled by `fang`. The goal was to gain polished help output, styled errors, man pages, and shell completions without swallowing the full `cmdguard` framework. The migration is **functionally complete and verified**: all tests pass, lint is clean, Nix builds, and the binary behaves correctly. However, several polish items remain — most notably making the `-C`/`--no-color` flag also suppress fang's own colors, and adding regression tests for the new CLI surface.

---

## a) FULLY DONE

### Code Migration

- [x] Replaced stdlib `flag` parsing with Cobra/pflag in `config.go`.
- [x] Added `upd.NewCommand(runE func(context.Context, *Config) error) (*cobra.Command, *Config)` as the canonical entry point.
- [x] Kept `upd.ParseFlags(args []string) (*Config, error)` for backwards compatibility and tests.
- [x] Bound all existing flags with short + long forms: `-q/--quiet`, `-n/--nop`, `--dry-run`, `-C/--no-color`, `-g/--greatest`, `-a/--all`, `-P/--pin-latest`, `--json`, `--verbose`, `-f/--file`, `-r/--registry`, `-c/--concurrency`, `--retries`, `-t/--timeout`.
- [x] Preserved `-V` short version flag alongside `--version` by manually registering a Cobra `version` flag with shorthand `V`.
- [x] Preserved the exact original multi-line version output (program, URL, description, copyright) via `cmd.SetVersionTemplate()`.
- [x] Wired `cmd/upd/main.go` to `fang.Execute(...)` with `fang.WithoutVersion()` and `fang.WithNotifySignal(SIGINT, SIGTERM)`.
- [x] Moved the full run pipeline into `executeRun(ctx, cfg)` so the signal-aware context comes from fang/Cobra.
- [x] Removed obsolete `PrintUsage()` and `PrintVersion()` functions from `config.go`.

### User-Facing Features

- [x] Styled `--help` output via fang with `Long` description, `Example` block, and color-coded flags/commands.
- [x] Styled error output via fang (`ERROR` header, usage hint for flag errors).
- [x] Hidden `man` command that emits roff man pages via `mango` + `roff`.
- [x] Hidden `completion` command (bash/zsh/fish) via Cobra, still functional but not shown in help.
- [x] Backwards-compatible `--noColor` hidden alias for `--no-color`.
- [x] `--dry-run` alias still maps to `--nop`.

### Tooling & Docs

- [x] Updated `.golangci.yml`:
  - Allowlisted `charm.land/fang/v2` in `cmd` depguard rule.
  - Allowlisted `github.com/spf13/cobra` and `github.com/spf13/pflag` in `main` depguard rule.
  - Excluded `github.com/spf13/cobra.Command` from `exhaustruct`.
- [x] Updated `flake.nix` `vendorHash` to `sha256-rUeYCJxMl55eZhPnSE2T/jMz0oQAZt0SB6LwuzQE8gg=`.
- [x] Updated `AGENTS.md` to describe the new Cobra/fang pipeline, dependency count, and signal handling.
- [x] Updated `docs/pro-contra-cmdguard-adoption.md` with a note that `fang` was adopted instead of `cmdguard`.
- [x] Updated `/home/lars/projects/cmdguard/docs/feedback/2026-07-16_upd-evaluation.md` to reflect the `fang` adoption.
- [x] `go.mod` remains at `go 1.26.4` (compatible with Nix build Go version).

### Verification

- [x] `GOEXPERIMENT=jsonv2 go test ./... -count=1` — PASS.
- [x] `GOEXPERIMENT=jsonv2 go test -race ./... -count=1` — PASS.
- [x] `GOEXPERIMENT=jsonv2 go vet ./...` — PASS.
- [x] `GOEXPERIMENT=jsonv2 golangci-lint run ./...` — 0 issues.
- [x] `nix run .#test` — PASS.
- [x] `nix run .#lint` — PASS.
- [x] `nix build .#default` — PASS.
- [x] Manual binary checks: `--help`, `--version`, `-V`, `man`, `completion bash`, dry-run against sample `package.json` — all work.

---

## b) PARTIALLY DONE

- [ ] **Unified color override** — The `-C`/`--no-color` flag controls `upd`'s internal table/progress/warning output, but fang's styled help and error rendering still uses `colorprofile` (which only respects `NO_COLOR` env var and TTY detection). Passing `-C` does not guarantee fang will be monochrome.
- [ ] **`--noColor` alias deprecation** — The old camelCase flag is hidden and still works, but there is no deprecation warning or documented timeline for removal.
- [ ] **Man page polish** — The generated roff still lists the hidden `--noColor` alias (mango does not honor Cobra's hidden flag flag). The short flags are rendered with `--` prefix in the roff (e.g., `--C --no-color`), which is a mango-cobra formatting quirk.
- [ ] **Completion discoverability** — The completion command is hidden from help, which is standard but means users must discover it via docs or shell completion setup guides.
- [ ] **Error punctuation** — fang's `DefaultErrorHandler` appends a period to `err.Error()`. Some `errorfamily` messages already end with punctuation, risking double periods in rare cases.
- [ ] **Usage line accuracy** — Fang renders `upd [--flags]` in the usage block, which is accurate but could be clearer about positional `[pattern ...]` args.

---

## c) NOT STARTED

### Features & UX

- [ ] Env-var support for flags (`UPD_REGISTRY`, `UPD_FILE`, `UPD_TIMEOUT`, etc.).
- [ ] Typo suggestions for unknown flags / subcommands (`did you mean --json?`).
- [ ] Custom `ColorSchemeFunc` that disables fang colors when `cfg.NoColor` is true.
- [ ] Structured logging or `--debug` log level.
- [ ] New subcommands: `check`, `doctor`, `init`, or `config`.
- [ ] Config file support (`.updrc`, `upd.json`, etc.).
- [ ] Human migration guide / blog post for the CLI change.
- [ ] Benchmark comparing old vs new binary size, startup time, and build time.
- [ ] Add `CHANGELOG.md` entry for this change.
- [ ] Commit the current working-tree changes.
- [ ] Add deprecation notice for `--noColor`.

### Tests

- [ ] Test for `man` command output (roff contains expected sections and flags).
- [ ] Test for `completion bash` output (starts with `# bash completion V2 for upd`).
- [ ] Test for version output format (multi-line template).
- [ ] Test for `--noColor` hidden alias still parsing.
- [ ] Test for `--no-color` canonical flag parsing.
- [ ] Test for styled error output (or at least exit code) on unknown flag.
- [ ] Test for signal handling via fang (hard but possible with a mock signal).
- [ ] Test for `NewCommand` returning a command with the correct `Use`/`Short`/`Long`.
- [ ] Test for `ParseFlags` returning `ErrHelp`/`ErrVersion` correctly.
- [ ] Test that `--version` and `-V` both return `ErrVersion` in `ParseFlags`.

### Docs & Maintenance

- [ ] Update `README.md` to mention the new styled help / man pages / completions.
- [ ] Update `docs/DOMAIN_LANGUAGE.md` if any CLI terminology changed.
- [ ] Re-render VHS demos (`nix run .#demo`) so published GIFs reflect the new help style.
- [ ] Add a Nix flake check for `go mod tidy` cleanliness to prevent `go` directive drift.
- [ ] Evaluate whether `PrintUsage`/`PrintVersion` removal breaks any external consumers.
- [ ] Consider splitting `cmd/upd/main.go` into smaller files if it grows further.
- [ ] Review fang dependency freshness and pin policy.
- [ ] Add `go mod verify` step to CI.

---

## d) TOTALLY FUCKED UP!

Nothing is catastrophically broken. The migration is green across all verification gates. However, the following are **material regressions / risks** that should be monitored:

1. **Binary size grew ~54%**: from ~7.1 MB to ~11.2 MB (Nix build). This violates the original "only 4 direct deps" / lean-CLI principle documented in `AGENTS.md`. The cost is accepted for the UX gains, but it is a regression in the "small binary" dimension.
2. **`go` directive fragility**: `go mod tidy` initially bumped `go.mod` to `go 1.26.5` because the local dev shell runs Go 1.26.5, but the Nix builder uses Go 1.26.4 and `GOTOOLCHAIN=local`, causing a build failure. I manually reverted to `go 1.26.4`, but this could drift again on the next `go mod tidy`.
3. **Error family preservation is subtle**: `fang.Execute` returns domain errors (e.g., `ErrRegistryUnavailable` = Transient/exit 75). I wrapped it with `errorfamily.Wrap(err, errorfamily.Classify(err), ...)` to preserve the family. If `Classify` ever misclassifies a Cobra parse error as Transient, exit codes will be wrong. This is currently correct but worth a regression test.
4. **Public API surface changed**: `PrintUsage` and `PrintVersion` are gone. Since `upd` is a single-binary module and these were in `package upd`, external consumers are unlikely, but this is technically a breaking change.

---

## e) WHAT WE SHOULD IMPROVE!

1. **Top priority: color scheme integration.** Make `-C`/`--no-color` disable fang's help/error colors, not just `upd`'s table output. This is the most visible user-facing inconsistency.
2. **Add CLI regression tests.** At minimum: `man`, `completion`, version format, `--noColor` alias, unknown flag error. These are cheap and prevent silent breakage.
3. **Add env-var support.** `UPD_REGISTRY` and `UPD_FILE` are low-hanging fruit and common in CI usage.
4. **Add typo suggestions.** Cobra doesn't do this; fang doesn't either. A small Levenshtein helper in `ParseFlags` would be a nice UX win.
5. **Fix man page hidden-flag leakage.** Either remove the `--noColor` alias entirely (breaking change) or find a way to hide it from mango's roff output.
6. **Document the change.** `README.md` and `CHANGELOG.md` should mention styled help, man pages, and completions.
7. **Automate `go mod tidy` guard.** Add a CI check that `go mod tidy` produces no diff, or pin the Nix builder to Go 1.26.5.
8. **Re-render VHS demos.** The published demos show the old hand-rolled help; they should show the new fang-styled help.
9. **Consider removing `--noColor` in v1.2.0.** With a deprecation warning in v1.1.x, we can drop the hidden alias in the next minor/major release.
10. **Benchmark the build.** Measure cold `nix build`, `go test`, and binary startup times before/after to quantify the dependency cost.

---

## f) Top 50 Things We Should Get Done Next

Sorted by a rough mix of user impact and engineering leverage:

1. Implement `ColorSchemeFunc` that disables fang colors when `cfg.NoColor` is true.
2. Add test for styled help output (or at least that it contains expected flags).
3. Add test for `man` command output.
4. Add test for `completion bash` output.
5. Add test for version output template.
6. Add test for unknown flag error and exit code.
7. Add env-var support for `UPD_REGISTRY`.
8. Add env-var support for `UPD_FILE`.
9. Add env-var support for `UPD_TIMEOUT`.
10. Add typo suggestions for unknown flags.
11. Add typo suggestions for unknown subcommands.
12. Remove `--noColor` from mango man page output (or remove alias).
13. Update `README.md` with new help/man/completion features.
14. Add `CHANGELOG.md` entry for fang/Cobra migration.
15. Re-render VHS demo GIFs with new help style.
16. Commit the current working-tree changes.
17. Add deprecation warning for `--noColor` alias.
18. Add `go mod tidy` cleanliness check to CI.
19. Add `go mod verify` to CI.
20. Update Nix builder to Go 1.26.5 (or use `GOTOOLCHAIN=auto`) to avoid `go` directive drift.
21. Add `cmd/upd` integration test that runs the binary end-to-end with mock registry.
22. Add test that `ParseFlags` returns `ErrHelp` for `-h` and `--help`.
23. Add test that `ParseFlags` returns `ErrVersion` for `-V` and `--version`.
24. Add test that `--noColor` alias still works.
25. Add test that `--no-color` canonical flag works.
26. Add test that `--dry-run` alias sets `cfg.Nop`.
27. Add test for signal cancellation via fang (mock SIGINT).
28. Review fang dependency update policy (v2 is new, watch for breaking changes).
29. Document `completion` command usage in `README.md`.
30. Document `man` command usage in `README.md`.
31. Add `upd --help` screenshot/example to `README.md`.
32. Consider renaming `NoColor` field to `DisableColor` for clarity.
33. Consider whether `Config` should be passed by value in `NewCommand` closure.
34. Add `TestNewCommand` verifying command metadata.
35. Add `TestBindFlags` table test covering all flags.
36. Add benchmark for `ParseFlags`.
37. Add benchmark for `NewCommand`.
38. Compare binary size in CI and alert on large increases.
39. Compare build time in CI and alert on large increases.
40. Add `nix flake check` to CI (currently only `build` + `test` + `lint` apps are used).
41. Review `cmd/upd/main.go` for further splitting if it grows.
42. Review whether `printWarnings` should respect `cfg.NoColor` (it already does via ANSI codes, but could use `Renderer`).
43. Investigate whether fang's error period appending can be disabled or customized.
44. Investigate if `--no-color` should imply `NO_COLOR` env var set for child processes.
45. Add a `doctor` command that checks registry reachability and `package.json` validity.
46. Add `--format` flag to select output format instead of separate `--json`.
47. Add `--silent` alias for `--quiet`.
48. Add `--update` alias for default behavior to be explicit.
49. Add issue templates for feature requests and bug reports in `.github/`.
50. Schedule a periodic dependency audit (e.g., monthly) given the new Charm ecosystem surface.

---

## g) Top 2 Questions I Cannot Figure Out Myself

1. **Should I commit the current fang/Cobra migration changes right now, or do you want to review the diff first?** The working tree is dirty with 8 modified files and no commit has been made for this session's work.

2. **Do you want me to implement the unified `-C`/`--no-color` behavior so it also disables fang's styled help and error colors, or should I leave it as a known limitation?** Doing it cleanly requires a `ColorSchemeFunc` closure that captures the parsed `Config.NoColor` after Cobra parses flags; it is straightforward but slightly increases `main.go` coupling.

---

## Appendix: Diagnostic Snapshot

- **LSP errors:** 0
- **LSP warnings:** 33 (all pre-existing `gopls` warnings about `encoding/json/v2` APIs requiring go1.27 while the module declares `go 1.26.4`; unrelated to this work).
- **golangci-lint issues:** 0
- **Test failures:** 0
- **Nix build:** success
- **Binary size:** ~11.2 MB (up from ~7.1 MB pre-migration)

---

## Files Modified This Session

```
 M .golangci.yml
 M AGENTS.md
 M cmd/upd/main.go
 M config.go
 M config_test.go
 M docs/pro-contra-cmdguard-adoption.md
 M flake.nix
 M go.mod
 M go.sum
```

Also updated outside the `upd` repo:

```
/home/lars/projects/cmdguard/docs/feedback/2026-07-16_upd-evaluation.md
```
