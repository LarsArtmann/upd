# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed

- **Documentation overhaul** (docs-health pass 2026-09-25) — created
  `ROADMAP.md`; refreshed `FEATURES.md` and `TODO_LIST.md` against the code;
  annotated and archived historical status reports; added Keep a Changelog
  compare links to this file.

### Fixed

_Nothing yet._

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

- **Dependency and toolchain refresh** — `go-atomic-write` v0.5.1,
  `go-error-family` v0.10.1, fang v2.0.1, lipgloss v2.0.6; Go directive
  bumped to 1.26.7; nixpkgs and flake inputs refreshed (`b6110bf`, `4e1c074`,
  `9cb5944`, `ebd9b1b`, `f020505`).
- **`errors.AsType` migration** — registry retry-error matching in `npm.go`
  moved to Go 1.26+ generic `errors.AsType` (`b6110bf`).
- **Documentation pass** — registry-client references renamed to `pnpm.go`
  across docs and user-facing examples switched from `npm` to `pnpm` CLI
  invocations (`7402f06`). Note: the source file itself remains `npm.go` —
  that commit renamed references only.
- **CI supply-chain hardening** — all GitHub Actions pinned to full commit
  SHAs; golangci-lint pinned to v2.12.2; depguard disabled; funlen widened;
  gosec G304/G115 excluded; Go sources formatted with dprint (`e13492e`).

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

[Unreleased]: https://github.com/LarsArtmann/upd/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/LarsArtmann/upd/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/LarsArtmann/upd/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/LarsArtmann/upd/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/LarsArtmann/upd/compare/2.9.2...v1.0.0
