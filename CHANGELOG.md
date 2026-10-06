# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- **Positional patterns restored through fang** — `upd react*` failed with
  `unknown command` since the fang migration registered hidden
  `man`/`completion` subcommands (cobra's default arg validation then
  rejected bare arguments on the root command). The root command now
  declares that positional arguments are patterns; a CLI-level end-to-end
  test pins the behavior.
- **`--format=table|json` replaces `--json`** — invalid values are rejected
  up front with a clear error. `--json` (and `UPD_JSON`) keep working but
  print a deprecation warning and are gone from help, man pages, and
  completions; typos like `--jso` now suggest `--format=json`.
- **`--silent` alias for `--quiet`** — registered as a first-class flag with
  the `-s` shorthand.
- **`.npmrc` support** — the registry URL and per-registry bearer tokens
  (`//host/:_authToken`) are read from the user's and the package file's
  `.npmrc` (project-local wins). Explicit `--registry`/`UPD_REGISTRY` still
  takes precedence; malformed or unsupported entries warn instead of
  silently breaking auth, and token values never appear in output.
- **Machine-readable error classification in `--json`** — error entries now
  carry `code` and `family` (e.g. `registry.package_not_found`,
  `rejection`) so CI consumers can act on error kinds instead of message
  text.
- **Deprecation warnings for invalid `UPD_*` env vars** — invalid values
  (e.g. `UPD_TIMEOUT=30`) warn instead of silently falling back to the
  default.
- **Typo suggestions for unknown flags** — `--jso` answers with
  `Did you mean --format=json?`, including when positional patterns
  accompany the flag.
- **Issue and pull request templates** (`bug_report.yml`,
  `feature_request.yml`, `PULL_REQUEST_TEMPLATE.md`).
- **Exit codes documented in `--help`** — the long description spells out the
  0/1/75 contract so the retry semantics are discoverable without the README.
- **Release automation** — `.goreleaser.yml` (Linux/macOS, amd64+arm64,
  `SHA256SUMS`) plus a tag-triggered `release.yml` workflow;
  `docs/RELEASING.md` documents the full ritual.

### Fixed

- **Usage errors exit 1 instead of 75** — flag typos and missing flag values
  were classified as transient (`75`, "safe to retry"), telling CI to retry
  invocations that can never succeed. Flag-parse failures are now classified
  as Rejection (`1`); registry-side transience is unaffected.
- **Data race in the progress bar** — concurrent fetches each wrote the
  progress line without synchronization, corrupting output (and failing the
  new local `-race` gate). Progress writes are now serialized.
- **Unknown-flag suggestions now fire with positional arguments** —
  `upd --jso -f package.json` previously reported `unknown command
  "package.json"` because cobra's command lookup failed before flag parsing.

### Changed

- **Every package failure is carried in the partial-failure error** —
  scripts can now `errors.Is`/`As` over the run error instead of parsing
  rendered output.
- **CI hardening** — `nix flake check` job added; `govulncheck` no longer
  allowed to fail (verified clean); `go mod tidy -diff` + `go mod verify`
  steps; test timeout of 120s; 80% coverage gate (coverage is 87.7%).
- **Local test gate matches CI** — `nix run .#test` now runs with `-race`.
- **Test coverage expanded** — CLI-level end-to-end tests (update, dry-run,
  partial failure, JSON output, quiet, positional patterns, `.npmrc` auth),
  a SIGINT cancellation test, property tests for the version regexes,
  compile-verified package example, and opt-in integration tests against
  the real NPM registry (`nix run .#test-integration`).

## [1.4.0] - 2026-10-05

### Fixed

- **Build break from the `gobwas/glob` v1.0.0 bump** — `a98c59a` upgraded
  `gobwas/glob` v0.2.3 → v1.0.0 without updating the code for the upstream API
  rename (`glob.Glob` interface → `*glob.Pattern`), leaving master
  uncompilable from 2026-09-15 until 2026-09-25. `manifest.go` updated to the
  `*Pattern` API.
- **Lint config schema** — wrapcheck key typo `ignore-sig-regex-es` corrected
  to `ignore-sig-regexps` (rejected by golangci-lint's schema validator);
  `//nolint:exhaustruct` directives migrated to `exhaustruct_v5` (the linter
  renamed, leaving 12 unsuppressed findings); stale `exhaustruct` exclusion
  entry removed.

### Changed

- **CI supply-chain hardening** — all GitHub Actions pinned to full commit SHAs;
  depguard disabled; funlen widened; gosec G304/G115 excluded; Go sources
  formatted with dprint (`e13492e`). golangci-lint action pinned to v2.14.0
  (v2.12.2 predates `exhaustruct_v5` support).
- **Dependency refresh** — `go-atomic-write` v0.6.0, `go-error-family` v0.11.0,
  `lipgloss` v2.0.6, `xo/terminfo` v1.2.0, `gobwas/glob` v1.0.0;
  `.github/dependabot.yml` added (`a98c59a`, `7a1e31f`).
- **Toolchain drift fix** — go 1.27 builder, `rev` version, nixfmt, build check
  (`d43f460`).
- **Documentation overhaul** (docs-health pass 2026-09-25) — created
  `ROADMAP.md`; refreshed `FEATURES.md` and `TODO_LIST.md` against the code;
  annotated and archived historical status reports; added Keep a Changelog
  compare links to this file.

## [1.3.0] - 2026-08-16

### Fixed

- **Zero-value `Config` deadlock** — `Concurrency=0` produced an unbuffered
  semaphore channel that deadlocked `FetchAll` before any worker could launch
  (root cause of a 5-minute hang in a downstream consumer). Three layers of
  defense: `Config.Validate()` clamps `Concurrency<=0`, `Timeout<=0`,
  `Retries<0`, and an empty `Registry`; `NewEngine` always validates so no
  caller can bypass it; the `FetchAll` semaphore send is context-aware.
  `NewRegistryClient` also clamps non-positive timeouts to the 20s default.
  New tests cover deadlock prevention, context cancellation, and config
  clamping. (`59bcb48`)

### Changed

- **Dependency refresh** — `go-atomic-write` v0.4.1 plus Charm-stack transitive
  bumps (ultraviolet, charmtone, go-colorful) and nixpkgs/flake input refresh
  (`4e1c074`, `9cb5944`, `ebd9b1b`, `f020505`).
- **Documentation pass** — registry-client references renamed to `pnpm.go`
  across docs and user-facing examples switched from `npm` to `pnpm` CLI
  invocations (`7402f06`). Note: the source file itself remains `npm.go` —
  that commit renamed references only.

## [1.2.0] - 2026-07-26

### Added

- **fang/Cobra CLI migration** — replaced stdlib `flag` with `charm.land/fang/v2` + Cobra. Adds styled `--help` output, styled error rendering, hidden `man` and `completion` commands, and signal-aware execution via `fang.WithNotifySignal`.
- **Unified no-color override** — `-C`/`--no-color` and `UPD_NO_COLOR` now disable fang's styled help and error colors in addition to `upd`'s own table/progress/warning output.
- **Environment variable support** — every public flag can be set via `UPD_*` env vars (e.g., `UPD_REGISTRY`, `UPD_FILE`, `UPD_TIMEOUT`). CLI flags override env vars; invalid env values fall back to defaults.
- **CLI regression tests** — coverage for `NewCommand` metadata, `--version`/`-V` template, `man` command roff output, `completion bash` output, `--noColor` alias, `--dry-run`, unknown flag errors, and env-var precedence.
- **`charm.land/lipgloss/v2` direct dependency** — used to implement the no-color `ColorSchemeFunc` for fang.

### Changed

- **Flag canonicalization** — `--no-color` is the canonical long form; `--noColor` remains a hidden backwards-compatible alias.
- **README expanded** — added documentation for styled help/man/completions, environment variables, and shell completion setup.
- **Direct dependency count increased** to 8 (added `charm.land/fang/v2` and `charm.land/lipgloss/v2` on top of the pre-migration dependencies).
- **Minimum Go toolchain** bumped to Go 1.26.5 (`GOEXPERIMENT=jsonv2` still required).

### Fixed

- `-C`/`--no-color` no longer leaves fang's help and error rendering colored when the flag is explicitly set.

## [1.1.0] - 2026-07-16

### Added

- **TOCTOU-safe atomic writes** — package.json is written via a temp file,
  fsync, fingerprint verification, and atomic rename. If another process
  (pnpm install, IDE formatter) modifies the file during upd's network-fetch
  window, the write is aborted with `ErrConcurrentModification` and the file
  is left untouched. Powered by `go-atomic-write` v0.2.0.
- **`--pinLatest` / `-P` flag** — pins dependencies using the bare `latest`
  dist-tag to their exact resolved semver version (e.g. `"latest"` → `"7.7.4"`).
- **`--json` output mode** — machine-readable JSON output for CI pipelines and
  scripting. Includes summary stats (updated/kept/errors/total), package list,
  and structured error details.
- **`--verbose` flag** — shows full error chains (`%+v`) in the error detail
  block for deep debugging.
- **`--registry` / `-r` flag** — use a custom or private NPM registry URL.
- **`--timeout` / `-t` flag** — per-request HTTP timeout (default: 20s).
- **`--retries` flag** — max retries for transient 429/5xx failures (default: 3)
  with exponential backoff (1s base, 30s cap) and `Retry-After` header support.
- **`--dry-run` alias** — alias for `--nop` / `-n`.
- **Registry error classification** — 404/410 → `ErrPackageNotFound` (permanent,
  exit 1); 5xx/timeout → `ErrRegistryUnavailable` (transient, exit 75). Lets CI
  scripts distinguish retryable from permanent failures.
- **`ErrPartialFailure`** — non-zero exit code (1) when any package fails to
  resolve. Successful updates are still written to disk before the error is
  returned.
- **`ErrConcurrentModification`** — aborts the write when the on-disk file
  fingerprint no longer matches what was read, preventing data loss.
- **SIGINT/SIGTERM cancellation** — signal-aware context cancels in-flight HTTP
  requests during the fetch phase for graceful shutdown.
- **Auto color detection** — `NO_COLOR` env var and non-TTY stdout automatically
  disable ANSI colors without requiring the `-C` flag.
- **`go-error-family` adoption** — structured error classification with typed
  families (Rejection, Transient, Corruption, Conflict), exit codes derived from
  `Family.ExitCode()`, and structured context attached at creation sites.
- **Terminal width detection** — progress bar respects `COLUMNS` env var
  (fallback: 80 chars).
- **HTTP transport tuning** — `MaxIdleConns=100`, `MaxIdleConnsPerHost=16`,
  `IdleConnTimeout=90s` for efficient connection reuse.
- **`RendererOptions` struct** — render configuration consolidated into a
  typed options struct.
- **Benchmark tests** — 14 benchmarks across diff, glob, manifest building,
  and version replacement.
- **Integration tests** — mock HTTP registry server exercising the full
  read → fetch → write pipeline, including scoped package URL encoding.
- **golangci-lint and govulncheck** — added to CI pipeline and devShell.
- **VHS animated demos** — rendered and published to `vhs.charm.sh` cloud.
- **`FEATURES.md`** — honest feature inventory by status.
- **`TODO_LIST.md`** — actionable improvement tasks with evidence.
- **`docs/atomic-writes.md`** — dedicated documentation for the atomic write
  mechanism.
- **`docs/DOMAIN_LANGUAGE.md`** — domain-driven design glossary.

### Changed

- **`encoding/json/v2` + `encoding/json/jsontext` migration** — replaced
  `tidwall/gjson` with standard library JSON for byte-precise surgical edits
  via `jsontext.Decoder` streaming. Requires `GOEXPERIMENT=jsonv2`.
- **Error handling overhaul** — 13 domain sentinel errors, per-spec error
  carriers (`Spec.Err`), error detail block in terminal output, and a warnings
  pipeline for non-fatal issues (malformed sections, invalid glob patterns).
- **License switched to MIT** with dual authors.
- **README rewritten** — expanded usage, troubleshooting section, exit codes
  table, and atomic-writes documentation.
- **Copyright year** updated to 2026.
- **`errors.AsType` adoption** — registry retry-error matching in `npm.go`
  moved to Go 1.26+ generic `errors.AsType` (`b6110bf`); nixpkgs/charmtone
  bumps from `b6110bf`/`7343d0f` in the same window. Added retroactively —
  omitted when v1.2.0 was cut (flagged by the release report, item f.24).
- Direct dependency count reduced to 4 (semver, glob, go-atomic-write,
  go-error-family) — `tidwall/gjson` removed.

### Fixed

- Swapped VERSION OLD / VERSION NEW columns in noColor (`-C`) mode — the
  diff-highlight fallback now returns the correct string for each column.
- `makezero`, `cyclop`, and `tagliatelle` lint warnings resolved.
- Zero jscpd code clones achieved across all test files via helper extraction.

## [1.0.0] - 2026-06-18

First stable release of the Go port.

### Added

- Complete Go port of the original JavaScript `upd` CLI
- Concurrent NPM registry queries with configurable connection pool (`-c`)
- Semantic version resolution: latest stable (`dist-tags.latest`) or greatest (`-g`)
- Formatting-preserving `package.json` editing via byte-level JSON patching
- Character-level diff highlighting with ANSI colors in the upgrade table
- Real-time progress bar during registry lookups
- Glob-based dependency filtering with positive and negative (`!`) patterns
- Support for all four dependency sections: `dependencies`, `devDependencies`,
  `peerDependencies`, `optionalDependencies`
- Embedded `upd` field in `package.json` for default CLI arguments
- Nix flake with `buildGoModule`, devShell, and `nix fmt` support
- GitHub Actions CI workflow
- Version injection at build time via ldflags
- Comprehensive test suite covering engine, rendering, diff, manifest, and
  package.json editing (race-detector clean)

### Changed

- Rewritten from JavaScript to Go for single-binary distribution and
  compile-time type safety

### Removed

- All original JavaScript source files

[Unreleased]: https://github.com/LarsArtmann/upd/compare/v1.4.0...HEAD
[1.4.0]: https://github.com/LarsArtmann/upd/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/LarsArtmann/upd/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/LarsArtmann/upd/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/LarsArtmann/upd/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/LarsArtmann/upd/compare/2.9.2...v1.0.0
