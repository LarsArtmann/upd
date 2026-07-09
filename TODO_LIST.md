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

| #   | Task                                                                      | Evidence                                                                                                 |
| --- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| D1  | Use Go `min()`/`max()` builtins in `progress.go`                          | `progress.go:41,43,49` — confirmed                                                                       |
| D2  | Tagged switch on `spec.State` in `render.go`                              | `render.go:152` — `switch spec.State`                                                                    |
| D3  | Share `http.Client` across goroutines                                     | `npm.go:24` — single client in `RegistryClient`                                                          |
| D4  | Write `flake.nix` with build/test/lint/run/demo apps                      | `flake.nix` — all apps present, with meta descriptions                                                   |
| D5  | Write project `AGENTS.md`                                                 | `AGENTS.md` — comprehensive, kept current                                                                |
| D6  | Version injection via `-ldflags`                                          | `config.go:24` — `ProgramVersion` var                                                                    |
| D7  | GitHub Actions CI (build + vet + race test)                               | `.github/workflows/ci.yml`                                                                               |
| D8  | `-P`/`--pin-latest` flag with tests                                       | `config.go`, `config_test.go`, `manifest_test.go`, `engine_test.go`                                      |
| D9  | `encoding/json/v2` migration (replaced `tidwall/gjson`)                   | Zero `gjson` refs; `go.mod` has 3 deps                                                                   |
| D10 | `makezero` lint warnings fixed                                            | `diff.go:44`, `render.go:167` — pre-allocated                                                            |
| D11 | Copyright year → 2026                                                     | `config.go:164`                                                                                          |
| D12 | `latestReplaceRe` consolidation                                           | Removed; `resolveSpecVersion` sets `SNew=VNew`                                                           |
| D13 | `FEATURES.md`                                                             | This file — updated with all new features                                                                |
| D14 | `cmd/upd/main_test.go` — exit code + warnings tests                       | 8 tests, all passing                                                                                     |
| D15 | `doc.go` example signatures fixed                                         | Updated to new `BuildManifest`/`GetUpdArgs` API + new Config fields                                      |
| D16 | AGENTS.md updated with error handling (Spec.Err, exit codes)              | Pipeline steps 4,7,8,9; Gotchas; Domain concepts                                                         |
| D17 | Error handling overhaul (`Spec.Err`, `ErrRegistryUnavailable`, warnings)  | 13 sentinels, exit 75, error detail block, warnings pipeline                                             |
| D18 | VHS demo rendered + published                                             | `demo/demo.gif` exists, README has cloud URL                                                             |
| D19 | CONTRIBUTING.md updated with GOEXPERIMENT + nix workflow                  | Done                                                                                                     |
| D20 | README usage line fixed (`-c <concurrency>`)                              | Done                                                                                                     |
| D21 | README install commands fixed (`GOEXPERIMENT=jsonv2`)                     | Done                                                                                                     |
| D22 | docs/DOMAIN_LANGUAGE.md rewritten with actual terms                       | Done                                                                                                     |
| D23 | Per-package errors affect exit code (`ErrPartialFailure` → exit 1)        | `errors.go`, `cmd/upd/main.go:finalizeRun` — default behavior, no flag                                   |
| D24 | Consolidate quiet/non-quiet fetch+apply duplication in `main.go`          | Single code path; reporter set to noop when quiet. `cmd/upd/main.go`                                     |
| D25 | Add golangci-lint to CI workflow                                          | `.github/workflows/ci.yml` — separate lint job via `golangci-lint-action`                                |
| D26 | Add `golangci-lint` to `nix run .#lint`                                   | `flake.nix` — lint app now runs `golangci-lint run ./...`                                                |
| D27 | Add integration test with mock HTTP registry server (full pipeline)       | `integration_test.go` — `TestFullPipelineReadFetchWrite`, `TestFullPipelineDryRunDoesNotWrite`           |
| D28 | Add HTTP retry logic for transient NPM registry failures (429, 5xx)       | `npm.go` — exponential backoff, Retry-After header. 6 tests in `npm_test.go`. `--retries` flag added.    |
| D29 | Add `--registry <url>` flag for custom/private NPM registry               | `config.go`, `npm.go:NewRegistryClient(cfg)`. Tested: `TestParseFlagsRegistryFlag`. `-r` short form.     |
| D30 | Add context deadline for entire fetch phase                               | `cmd/upd/main.go` — `signal.NotifyContext` for SIGINT/SIGTERM cancellation. Per-request via `--timeout`. |
| D31 | Auto-detect non-TTY and disable colors (`NO_COLOR` env var, isatty check) | `config.go:ShouldDisableColor`. Auto-applied in `run()` when `-C` not set.                               |
| D32 | Add `--dry-run` as alias for `-n`/`--nop`                                 | `config.go` — `--dry-run` registered as alias. Tested: `TestParseFlagsDryRunAlias`.                      |
| D33 | Add `--timeout <seconds>` flag                                            | `config.go` — `-t`/`--timeout` (default: 20s). Tested: `TestParseFlagsTimeoutFlag`.                      |
| D34 | Add `--json` output mode for CI/scripting                                 | `render.go:RenderJSON`. Tested: 3 tests in `render_json_test.go`.                                        |
| D35 | Verify scoped package URL encoding (`@scope/name`) against live registry  | `integration_test.go:TestScopedPackageURLEncoding` — verified against mock registry.                     |
| D36 | Run `govulncheck` and fix findings                                        | CI has vulncheck job. 1 finding: stdlib `crypto/tls` vuln (GO-2026-5856) — needs Go 1.26.5 toolchain.    |
| D37 | Document exit codes in `--help` output + README                           | `config.go:PrintUsage` shows exit codes. README has Exit Codes table + Troubleshooting section.          |
| D38 | Add "Troubleshooting" section to README                                   | README — covers 404, registry down, concurrent mod, invalid JSON, progress bar, colors.                  |
| D39 | Decide quiet-mode + warnings interaction                                  | `-q` now suppresses warnings too. Documented in `--help` and README.                                     |
| D40 | Add `--verbose` flag for full error chains                                | `config.go`, `render.go` — `--verbose` uses `%+v` formatting in error detail block.                      |
| D41 | Add bench tests for diff, glob, manifest building                         | `benchmark_test.go` — 14 benchmarks across diff, patterns, manifest, replaceVersion.                     |
| D42 | Detect terminal width for progress bar (instead of hardcoded 80)          | `progress.go:clearWidth` — checks `COLUMNS` env var, falls back to 80.                                   |
| D43 | Tune HTTP transport (MaxIdleConns, IdleConnTimeout)                       | `npm.go` — MaxIdleConns=100, MaxIdleConnsPerHost=16, IdleConnTimeout=90s. Named constants.               |
| D44 | Add `meta.description` to all nix apps                                    | `flake.nix` — all 5 apps have descriptions.                                                              |
| D45 | Add govulncheck to CI                                                     | `.github/workflows/ci.yml` — vulncheck job runs `govulncheck ./...`.                                     |
| D46 | Add golangci-lint + govulncheck to devShell                               | `flake.nix` — devShell now includes `golangci-lint` and `govulncheck`.                                   |

---

## NOT DONE — Medium Priority

| #   | Task                                                          | Source        | Notes                                                                                 |
| --- | ------------------------------------------------------------- | ------------- | ------------------------------------------------------------------------------------- |
| 47  | Add `.npmrc` parsing for custom registry config               | e11, #22      | `--registry` flag covers the primary use case. `.npmrc` would add auth token support. |
| 48  | Add release automation (GoReleaser or tag-based pipeline)     | #9, #13       | No release workflow                                                                   |
| 49  | Add Renovate/Dependabot config                                | #24, f38      | No dependency automation                                                              |
| 50  | Add `nix flake check` to CI                                   | #37           | Not in CI                                                                             |
| 51  | Add coverage threshold to CI (fail if <80%)                   | #23           | Not in CI                                                                             |
| 52  | Add shell completions (bash/zsh/fish)                         | #20, #47, f28 | Not implemented                                                                       |
| 53  | Add man page (`man/upd.1`)                                    | #48, f41      | Not implemented                                                                       |
| 54  | Add property-based tests for `versionRe` and `latestRe` regex | f18           | Not implemented                                                                       |
| 55  | Add Go doc examples with `// Output:` to `doc.go`             | f13           | Example exists but not compile-tested                                                 |
| 56  | Consider `errors.Join` for multi-error aggregation            | err20         | Currently N separate warnings                                                         |
| 57  | Add focused demo tapes (`pin-latest.tape`, `greatest.tape`)   | f11           | Only one tape exists                                                                  |
| 58  | Add integration test (build-tagged) hitting real NPM registry | f12           | All tests use mocks                                                                   |
| 59  | Add issue/PR templates to `.github/`                          | f45           | Not implemented                                                                       |
| 60  | Review all error messages for user-facing quality             | err37, f37    | Not audited                                                                           |
| 61  | Add `slog` structured logging                                 | e15, err28    | No logging stack                                                                      |

## REJECTED (with reasoning)

| #   | Task                                                 | Reason                                                                                                 |
| --- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| R1  | Dockerfile                                           | Single static binary makes Docker unnecessary.                                                         |
| R2  | `BuildManifest` options struct                       | YAGNI — no external library consumers; positional bool is fine.                                        |
| R3  | `Spec.Section` typed enum                            | YAGNI — bare string works; no bugs from it.                                                            |
| R4  | `PackageName` branded type                           | YAGNI — adds ceremony without preventing real bugs.                                                    |
| R5  | Golden file tests vs `rse/upd`                       | Go port has different output format; byte-for-byte parity is artificial.                               |
| R6  | Adopt go-error-family / oops / bridge                | See `docs/research/2026-07-09_error-handling-libraries.md`. stdlib is sufficient.                      |
| R7  | `sjson` for writes                                   | Current `jsontext.Decoder` byte-splice approach works and is tested.                                   |
| R8  | Surface `ErrRegistryUnavailable` in non-fatal path   | Partial failure mixes 404 + 503; exit 1 is correct (D23). Exit 75 reserved for total registry failure. |
| R9  | Additional exit codes (65=EX_DATAERR, 66=EX_NOINPUT) | Only 0, 1, 75 used; adding more adds complexity without clear value.                                   |
| R10 | `--fail-on-error` flag                               | Resolved by D23 — non-zero exit is the default behavior; no flag needed.                               |
