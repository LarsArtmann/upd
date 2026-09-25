# TODO List

> Short- and mid-term improvement tasks, verified against the actual codebase.
> Harvested from status reports (`docs/status/`), release audits, and code reads.
> De-duplicated. **Open items only** — completed work lives in `CHANGELOG.md`.

---

## Release completion (do first)

| # | Task                                                                                                                                                                                                                       | Source                                   | Notes                                                                                                                              |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **Complete the v1.3.0 release**: create the GitHub Release for the existing `v1.3.0` tag, and fix the version source — `flake.nix:22` still says `1.2.0`, so every binary built from source or the tag reports `upd 1.2.0` | tag `v1.3.0` → `59bcb48`; `flake.nix:22` | Tag is pushed; no GitHub Release exists (`gh release list` shows v1.2.0 as Latest). Discovered in the 2026-09-25 docs-health pass. |
| 2 | Add `nix flake check` to CI and to the pre-release verification gate                                                                                                                                                       | 2026-07-26 report c.3, f.25              | Still absent from `.github/workflows/ci.yml`.                                                                                      |
| 3 | Release automation: `release.yml` on tag push, goreleaser prebuilt binaries (Linux/macOS amd64+arm64), `SHA256SUMS` + signatures                                                                                           | 2026-07-26 report f.9–f.10, f.18         | No release workflow, no prebuilt binaries; install currently requires Nix or Go 1.26+ with `GOEXPERIMENT=jsonv2`.                  |
| 4 | Write `docs/RELEASING.md` runbook (bump → changelog → tag → push → release, exact commands)                                                                                                                                | 2026-07-26 report f.6                    | Every release is currently a reverse-engineered manual ritual.                                                                     |
| 5 | Document the stale `2.x` git tags (2020–2023, JS-era) as legacy artifacts not part of the `v1.x` line                                                                                                                      | 2026-07-26 report f.19–f.21              | They sort above `v1.x` in plain `git tag` output. (Deleting them is a separate user decision — irreversible.)                      |
| 6 | Verify govulncheck is clean on Go 1.26.7 (GO-2026-5856 was fixed in 1.26.5), then remove `continue-on-error` from the CI vulncheck job                                                                                     | `ci.yml:52`; 2026-07-09_18-55 e.15       | `ci.yml:52` still masks the whole job.                                                                                             |

## CLI & UX

| #  | Task                                                                                                                   | Source                               | Notes                                                             |
| -- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | ----------------------------------------------------------------- |
| 7  | Deprecation warning for `--noColor` alias + removal plan; fixes the man-page leak (mango renders the hidden alias)     | 2026-07-16_07-01 b/d.2–d.3           | Top-priority leftover of the fang migration.                      |
| 8  | Warn (don't silently fall back) when an `UPD_*` env var is set but invalid (e.g. `UPD_TIMEOUT=30` keeps default `20s`) | 2026-07-16_07-01 b/d.1               | `applyEnvFlags` silently ignores bad values (`config.go`).        |
| 9  | Typo suggestions for unknown flags (`did you mean --json?`)                                                            | 2026-07-16_05-30 e.4                 | Small Levenshtein helper in flag parsing.                         |
| 10 | Re-render VHS demos (`nix run .#demo`) so published GIFs show the fang-styled help                                     | 2026-07-16_07-01 c                   | README GIF predates the CLI restyle.                              |
| 11 | `--format` flag (table/json) instead of separate `--json`; optional `--silent` alias for `--quiet`                     | 2026-07-16_05-30 f.46–f.47           |                                                                   |
| 12 | `.npmrc` parsing: registry URL + auth token support                                                                    | 2026-06-17_18-26 f.22; prior TODO #1 | `--registry` covers the URL; `.npmrc` adds private-registry auth. |
| 13 | Surface errorfamily `code`/`family` in `--json` output (additive fields) and leverage `Format('+')` in `--verbose`     | 2026-07-16_00-50 f.9–f.10            | Machine-readable error codes for CI consumers.                    |

## Testing

| #  | Task                                                                                                          | Source                                 | Notes                                                       |
| -- | ------------------------------------------------------------------------------------------------------------- | -------------------------------------- | ----------------------------------------------------------- |
| 14 | End-to-end test of `cmd/upd` `run()` with a mock registry (full pipeline: fetch → render → write → exit code) | 2026-07-16_00-50 c.2, e.2–e.3          | `run()`/`finalizeRun()` still have no direct tests.         |
| 15 | Signal-handling test: mock SIGINT through fang, verify fetch cancellation                                     | 2026-07-09_18-55 b; 2026-07-16_05-30 c |                                                             |
| 16 | Property-based tests for `versionRe` / `latestRe` regexes                                                     | prior TODO #8                          |                                                             |
| 17 | Compile-tested doc examples in `doc.go` (`// Output:`)                                                        | prior TODO #9                          | Example was fixed in `9ca148e` but is not compile-verified. |
| 18 | Coverage threshold gate in CI (fail if < 80%)                                                                 | prior TODO #5                          |                                                             |
| 19 | Derive `updates` count from the manifest inside `RenderJSON` (drop the redundant param, like `errCount` was)  | 2026-07-09_18-55 e.3, f.9              |                                                             |
| 20 | Migrate benchmarks from `b.N` to `b.Loop()`                                                                   | 2026-07-09_18-55 e.11                  | Go 1.24+ pattern; gopls modernize hint.                     |
| 28 | Include `-race` in `nix run .#test` (or add a `.#test-race` app) so the local gate matches CI                 | 2026-07-26 report f.38                 | CI runs `-race`; the local test app currently doesn't.      |

## Maintenance & CI

| #  | Task                                                                      | Source                | Notes                                                     |
| -- | ------------------------------------------------------------------------- | --------------------- | --------------------------------------------------------- |
| 21 | `go mod tidy` cleanliness + `go mod verify` checks in CI                  | 2026-07-16_07-01 c    | Prevents `go` directive drift breaking the Nix build.     |
| 22 | `-timeout 120s` on the CI test step                                       | 2026-07-09_18-55 e.17 | A hung test currently runs until GitHub's 6h job timeout. |
| 23 | Issue/PR templates in `.github/`                                          | prior TODO #13        |                                                           |
| 24 | `errors.Join` for multi-error aggregation (N warnings → one joined error) | prior TODO #10        |                                                           |
| 25 | Focused demo tapes: `pin-latest.tape`, `greatest.tape`, `patterns.tape`   | prior TODO #11        |                                                           |
| 26 | Build-tagged integration test against the real NPM registry               | prior TODO #12        | All current tests use mocks.                              |
| 27 | `slog` structured logging                                                 | prior TODO #15        |                                                           |

---

## REJECTED (with reasoning)

| #   | Task                                                 | Reason                                                                                                                           |
| --- | ---------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| R1  | Dockerfile                                           | Single static binary makes Docker unnecessary.                                                                                   |
| R2  | `BuildManifest` options struct                       | YAGNI — no external library consumers; positional bool is fine.                                                                  |
| R3  | `Spec.Section` typed enum                            | YAGNI — bare string works; no bugs from it.                                                                                      |
| R4  | `PackageName` branded type                           | YAGNI — adds ceremony without preventing real bugs.                                                                              |
| R5  | Golden file tests vs `rse/upd`                       | Go port has different output format; byte-for-byte parity is artificial.                                                         |
| R6  | `sjson` for writes                                   | Current `jsontext.Decoder` byte-splice approach works and is tested.                                                             |
| R7  | Surface `ErrRegistryUnavailable` in non-fatal path   | Partial failure mixes 404 + 503; exit 1 is correct. Exit 75 reserved for total registry failure.                                 |
| R8  | Additional exit codes (65=EX_DATAERR, 66=EX_NOINPUT) | Only 0, 1, 75 used; adding more adds complexity without clear value.                                                             |
| R9  | `--fail-on-error` flag                               | Resolved — non-zero exit is the default behavior (`ErrPartialFailure`); no flag needed.                                          |
| R10 | `upd update` self-update command                     | Go binaries don't self-update; installs are managed by nix/go install. Same reasoning as the rejected update-notifier.           |
| R11 | Deleting the legacy `2.x` git tags                   | Irreversible, and nothing breaks from their existence once documented (TODO #5). Revisit only if tooling actually trips on them. |
