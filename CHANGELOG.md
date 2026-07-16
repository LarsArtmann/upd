# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [1.1.0] - 2026-07-16

### Added

- **TOCTOU-safe atomic writes** — package.json is written via a temp file,
  fsync, fingerprint verification, and atomic rename. If another process
  (npm install, IDE formatter) modifies the file during upd's network-fetch
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
