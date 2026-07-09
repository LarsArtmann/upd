# TODO List

> Short- and mid-term improvement tasks, verified against the actual codebase.
> Derived from status reports, research docs, and code audits. De-duplicated.
> Items marked DONE are kept for traceability — remove after each release.

---

## Files reviewed for TODOs

- [x] `docs/status/2026-06-17_18-03_go-port-complete.md`
- [x] `docs/status/2026-06-17_18-26_round2-quality-and-devops.md`
- [x] `docs/status/2026-06-28_04-14_post-atomic-write-upgrade.html`
- [x] `docs/status/2026-07-09_07-16_pinlatest-migration-and-review.md`
- [x] `docs/status/2026-07-09_09-08_jsonv2-vhs-migration-and-cleanup.md`
- [x] `docs/status/2026-07-09_14-52_error-handling-overhaul.md`
- [x] `docs/research/2026-07-09_error-handling-libraries.md`
- [x] `AGENTS.md`, `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `FEATURES.md`

---

## DONE (verified against code)

| #   | Task                                                                     | Evidence                                                               |
| --- | ------------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| D1  | Use Go `min()`/`max()` builtins in `progress.go`                         | `progress.go:41,43,49` — confirmed                                     |
| D2  | Tagged switch on `spec.State` in `render.go`                             | `render.go:120` — `switch spec.State`                                  |
| D3  | Share `http.Client` across goroutines                                    | `npm.go:24` — single client in `RegistryClient`                        |
| D4  | Write `flake.nix` with build/test/lint/run/demo apps                     | `flake.nix` — all apps present                                         |
| D5  | Write project `AGENTS.md`                                                | `AGENTS.md` — comprehensive, kept current                              |
| D6  | Version injection via `-ldflags`                                         | `config.go:24` — `ProgramVersion` var                                  |
| D7  | GitHub Actions CI (build + vet + race test)                              | `.github/workflows/ci.yml`                                             |
| D8  | `-P`/`--pin-latest` flag with tests                                      | `config.go:75`, `config_test.go`, `manifest_test.go`, `engine_test.go` |
| D9  | `encoding/json/v2` migration (replaced `tidwall/gjson`)                  | Zero `gjson` refs; `go.mod` has 3 deps                                 |
| D10 | `makezero` lint warnings fixed                                           | `diff.go:44`, `render.go:167` — pre-allocated                          |
| D11 | Copyright year → 2026                                                    | `config.go:164`                                                        |
| D12 | `latestReplaceRe` consolidation                                          | Removed; `resolveSpecVersion` sets `SNew=VNew`                         |
| D13 | `FEATURES.md`                                                            | This file                                                              |
| D14 | `cmd/upd/main_test.go` — exit code + warnings tests                      | 7 tests, all passing                                                   |
| D15 | `doc.go` example signatures fixed                                        | Updated to new `BuildManifest`/`GetUpdArgs` API                        |
| D16 | AGENTS.md updated with error handling (Spec.Err, exit codes)             | Pipeline steps 4,7,8,9; Gotchas; Domain concepts                       |
| D17 | Error handling overhaul (`Spec.Err`, `ErrRegistryUnavailable`, warnings) | 12 sentinels, exit 75, error detail block, warnings pipeline           |
| D18 | VHS demo rendered + published                                            | `demo/demo.gif` exists, README has cloud URL                           |
| D19 | CONTRIBUTING.md updated with GOEXPERIMENT + nix workflow                 | Done                                                                   |
| D20 | README usage line fixed (`-c <concurrency>`)                             | Done                                                                   |
| D21 | README install commands fixed (`GOEXPERIMENT=jsonv2`)                    | Done                                                                   |
| D22 | docs/DOMAIN_LANGUAGE.md rewritten with actual terms                      | Done                                                                   |
| D23 | Per-package errors affect exit code (`ErrPartialFailure` → exit 1)       | `errors.go`, `cmd/upd/main.go:finalizeRun` — default behavior, no flag |

---

## NOT DONE — High Priority

| #   | Task                                                                                          | Source                | Notes                                                                          |
| --- | --------------------------------------------------------------------------------------------- | --------------------- | ------------------------------------------------------------------------------ |
| 1   | Consolidate quiet/non-quiet fetch+apply duplication in `main.go`                              | e6, f2, #18           | `FetchAll` called twice (lines ~49 and ~59); `finalizeRun` partially extracted |
| 2   | Add golangci-lint to CI workflow                                                              | #10, #21, #34, f6, f7 | `.golangci.yml` exists with 100+ linters but CI runs only `go vet`             |
| 3   | Add `golangci-lint` to `nix run .#lint`                                                       | f7                    | Currently `go vet && go build` only                                            |
| 4   | Add integration test with mock HTTP registry server (full pipeline: read→fetch→compare→write) | e4, f3                | Engine tested but not with `httptest.Server` mock                              |
| 5   | Add HTTP retry logic for transient NPM registry failures (429, 5xx)                           | e10, #2, #40          | No retry/backoff code exists                                                   |

## NOT DONE — Medium Priority

| #      | Task                                                                      | Source               | Notes                                              |
| ------ | ------------------------------------------------------------------------- | -------------------- | -------------------------------------------------- |
| 6      | Add `--registry <url>` flag for custom/private NPM registry               | e11, #5, #41         | Registry hardcoded to `registry.npmjs.org`         |
| 7      | Add context deadline for entire fetch phase                               | e9, #3, #7, #10      | Only per-request 20s timeout; no overall deadline  |
| 8      | Auto-detect non-TTY and disable colors (`NO_COLOR` env var, isatty check) | e13, #4, #50, f8     | Only manual `-C` flag                              |
| 9      | Add `--dry-run` as alias for `-n`/`--nop`                                 | #17, #43, f27        | Conventional name; trivial to add                  |
| 10     | Add `--timeout <seconds>` flag                                            | #15, #24, #42        | Hardcoded 20s                                      |
| 11     | Add `--json` output mode for CI/scripting                                 | #12, #22, #45, f29   | Not implemented                                    |
| ~~12~~ | ~~Add `--fail-on-error` flag~~ — **RESOLVED by D23** (default behavior)   | err10                | Made non-zero exit the default; no flag needed     |
| 13     | Verify scoped package URL encoding (`@scope/name`) against live registry  | e12, #14, #16, #44   | `url.PathEscape` used but unverified               |
| 14     | Run `govulncheck` and fix findings                                        | #10, f23             | Never run                                          |
| 15     | Run `gosec` and fix findings                                              | #11                  | Never run                                          |
| 16     | Document exit codes in `--help` output + README                           | err16, err17         | Exit 75 undocumented in CLI help                   |
| 17     | Add "Troubleshooting" section to README                                   | err24                | Common errors (404, registry down, malformed JSON) |
| 18     | Decide quiet-mode + warnings interaction                                  | err11                | Should `-q` suppress `WARNING:` lines?             |
| 19     | Add `.npmrc` parsing for custom registry config                           | e11, #22             | No `.npmrc` support                                |
| 20     | Add bench tests for diff, glob, manifest building                         | e5, #6, #9, f11, f17 | No `Benchmark*` functions exist                    |

## NOT DONE — Low Priority

| #   | Task                                                                                     | Source        | Notes                                                                                                               |
| --- | ---------------------------------------------------------------------------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------- |
| 21  | Detect terminal width for progress bar (instead of hardcoded 80)                         | e7, #4        | `terminalResetWidth = 80` hardcoded                                                                                 |
| 22  | Tune HTTP transport (MaxIdleConns, IdleConnTimeout)                                      | #5, #16       | Default transport used                                                                                              |
| 23  | Add release automation (GoReleaser or tag-based pipeline)                                | #9, #13       | No release workflow                                                                                                 |
| 24  | Add Renovate/Dependabot config                                                           | #24, f38      | No dependency automation                                                                                            |
| 25  | Add `nix flake check` to CI                                                              | #37           | Not in CI                                                                                                           |
| 26  | Add coverage threshold to CI (fail if <80%)                                              | #23           | Not in CI                                                                                                           |
| 27  | Add shell completions (bash/zsh/fish)                                                    | #20, #47, f28 | Not implemented                                                                                                     |
| 28  | Add man page (`man/upd.1`)                                                               | #48, f41      | Not implemented                                                                                                     |
| 29  | Add property-based tests for `versionRe` and `latestRe` regex edge cases                 | f18           | Not implemented                                                                                                     |
| 30  | Add Go doc examples with `// Output:` to `doc.go`                                        | f13           | Example exists but not compile-tested                                                                               |
| 31  | Consider `errors.Join` for multi-error aggregation                                       | err20         | Currently N separate warnings                                                                                       |
| 32  | Surface `ErrRegistryUnavailable` in non-fatal path (per-package errors → exit 75)        | err13         | **REJECTED:** partial failure mixes 404 + 503; exit 1 is correct (D23). Exit 75 reserved for total registry failure |
| 33  | Add `--verbose` flag for full error chains                                               | err12         | `%+v` formatting of `spec.Err`                                                                                      |
| 34  | Consider additional exit codes (65=EX_DATAERR, 66=EX_NOINPUT)                            | err14, err15  | Only 0, 1, 75 used                                                                                                  |
| 35  | Add `meta.description` to all nix apps                                                   | f15           | `nix flake check` warns about missing descriptions                                                                  |
| 36  | Add focused demo tapes (`pin-latest.tape`, `greatest.tape`, `patterns.tape`)             | f11           | Only one tape exists                                                                                                |
| 37  | Add integration test (build-tagged) hitting real NPM registry                            | f12           | All tests use mocks                                                                                                 |
| 38  | Add issue/PR templates to `.github/`                                                     | f45           | Not implemented                                                                                                     |
| 39  | Review all error messages for user-facing quality (What/Reassure/Why/Fix/Escape pattern) | err37, f37    | Not audited                                                                                                         |
| 40  | Add `slog` structured logging                                                            | e15, err28    | No logging stack                                                                                                    |

## REJECTED (with reasoning)

| #   | Task                                  | Reason                                                                            |
| --- | ------------------------------------- | --------------------------------------------------------------------------------- |
| R1  | Dockerfile                            | Single static binary makes Docker unnecessary.                                    |
| R2  | `BuildManifest` options struct        | YAGNI — no external library consumers; positional bool is fine.                   |
| R3  | `Spec.Section` typed enum             | YAGNI — bare string works; no bugs from it.                                       |
| R4  | `PackageName` branded type            | YAGNI — adds ceremony without preventing real bugs.                               |
| R5  | Golden file tests vs `rse/upd`        | Go port has different output format; byte-for-byte parity is artificial.          |
| R6  | Adopt go-error-family / oops / bridge | See `docs/research/2026-07-09_error-handling-libraries.md`. stdlib is sufficient. |
| R7  | `sjson` for writes                    | Current `jsontext.Decoder` byte-splice approach works and is tested.              |
