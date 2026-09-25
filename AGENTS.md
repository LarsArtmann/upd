# UPD — Agent Context

Enduring context for AI sessions working on `upd`.
User-facing description lives in `README.md`; domain glossary in `docs/DOMAIN_LANGUAGE.md`.

## What This Is

Go CLI that bumps dependency versions in an NPM `package.json` while **byte-preserving all original JSON formatting** (whitespace, key order, quotes). Only the concrete version inside a constraint string changes; everything else is untouched.

## Build / Test / Lint

This repo standardizes on **Nix flakes** — no Makefile, no justfile.

```bash
nix run .#build          # build to bin/upd
nix run .#test           # go test ./... -v -count=1
nix run .#lint           # go vet + go build + golangci-lint
nix run .#run -- <args>  # go run ./cmd/upd <args>
nix run .#demo           # render VHS demo GIFs locally
nix run .#demo -- --publish  # render + publish to vhs.charm.sh
```

**`GOEXPERIMENT=jsonv2` is required** for all Go commands (set automatically in nix apps and devShell; must be `export`ed for plain Go commands).

Plain Go equivalents: `GOEXPERIMENT=jsonv2 go build ./cmd/upd`, `GOEXPERIMENT=jsonv2 go test ./...`.
CI (`.github/workflows/ci.yml`) runs three jobs on push/PR to `master`: build + vet + race test, golangci-lint (pinned v2.12.2), and govulncheck (`continue-on-error: true` — see TODO_LIST #6). All actions are pinned to full commit SHAs.

Go sources are formatted with **dprint** (`dprint.json`, applied repo-wide in `e13492e`); there is no automated dprint gate yet. `nix fmt` formats only Nix files (nixpkgs-fmt).

VHS demos (`demo/*.tape`) are rendered with `vhs` and published to the [VHS cloud](https://vhs.charm.sh). GIFs are git-ignored — re-render with `nix run .#demo` anytime without committing binary files.

## Conventions

- **Nix-first**: all task automation lives in `flake.nix`. Do not add a Makefile or justfile.
- **Strict linting**: `.golangci.yml` enables 100+ linters (errcheck, wrapcheck, varnamelen, exhaustruct, ...). Expect loud diagnostics on untouched files — match surrounding style rather than chasing every pre-existing warning. `depguard` was **disabled** in `e13492e` (it used to restrict non-stdlib imports per directory); don't re-enable it without a decision.
- **Single root package**: all library code is package `upd` at the repo root; `cmd/upd` is the only executable.
- **git-town**: `git-town.toml` configures the branch workflow.

## Execution Pipeline

`cmd/upd/main.go:run()` drives a linear flow via `charm.land/fang/v2` + Cobra; the files below each own one stage:

1. **Build command** → `upd.NewCommand()` returns a `*cobra.Command` with flags bound to a `Config` (`config.go`). Env vars (`UPD_*`) are applied to their flags before CLI parsing, so explicit flags still override env values (`config.go:applyEnvFlags`). `fang.Execute()` styles help/errors with a `ColorSchemeFunc` that disables fang colors when `cfg.NoColor` is true, and adds hidden `man`/`completion` commands. Auto color detection via `ShouldDisableColor` (`NO_COLOR` env var, non-TTY stdout). The signal-aware context is supplied by `fang.WithNotifySignal()`.
2. **Parse flags** → Cobra/pflag parses short and long forms (`-h`/`--help`, `-c`/`--concurrency`, `-P`/`--pin-latest`, `-r`/`--registry`, `-t`/`--timeout`, `--retries`, `--json`, `--verbose`, `--dry-run`). `--no-color` is the canonical long form; `--noColor` remains a hidden alias for backwards compatibility. `ParseFlags()` is retained for tests.
3. **Read `package.json`** → `PackageFile` keeps the raw bytes and an xxhash64 fingerprint of them (`packagejson.go`). The fingerprint guards the later write against concurrent modifications.
4. **Merge patterns**: if `package.json` has an `upd` field (string or array), those args are **prepended** to CLI patterns.
5. **Build manifest** → `(Manifest, []string)` — the manifest is `name → []*Spec` across four sections in fixed order: `optionalDependencies`, `peerDependencies`, `devDependencies`, `dependencies` (`manifest.go`). The second return value is a slice of warning strings (malformed sections, invalid glob patterns). Takes `pinLatest bool` as third arg. `GetUpdArgs()` and `GetDependencySection()` return `error` as second value.
6. **Classify** each spec into a `State` via the version regex (`versionRe`) + `latestRe` + glob pattern matching.
7. **Fetch** packuments concurrently (semaphore bounded by `-c`, default 8) from registry (`engine.go`, `npm.go`). Names are lowercased before fetch. `NewEngine` runs `Config.Validate()` first, which clamps `Concurrency<=0`, `Timeout<=0`, `Retries<0`, and an empty `Registry` to defaults — a zero-value `Config` used to deadlock `FetchAll` (`59bcb48`). `RegistryClient` takes `*Config` (supports custom registry URL, timeout, retries). Retry logic: 429/5xx retried with exponential backoff + `Retry-After` header support. HTTP transport tuned (MaxIdleConns, IdleConnTimeout). Context is signal-aware (SIGINT/SIGTERM cancellation).
8. **Apply updates**: resolve target version, compare semver, mutate `PackageFile` bytes in place. When `IsLatest=true`, `SNew` is set directly. Errored specs carry their concrete error in `Spec.Err`.
9. **Render** terminal table (`render.go`) or JSON output (`--json`, `RenderJSON`). Error detail block supports `--verbose`. **Write back** only if updates occurred and `-n` is not set. Atomic write via `go-atomic-write`.
10. **Exit code** (`cmd/upd/main.go`): `ErrRegistryUnavailable` → exit 75; `ErrPartialFailure` → exit 1; all other errors → exit 1; success → exit 0. Warnings print to stderr but are suppressed in quiet mode.

## Domain Concepts

- **Packument**: NPM registry JSON for one package (`dist-tags`, `versions`, ...). Held as raw bytes in `npm.go`.
- **Manifest**: `map[string][]*Spec` — every occurrence of a dependency across all sections.
- **Spec**: one dependency in one section. Fields: `Section`, `Name`, `SOld`/`SNew` (full constraint string), `VOld`/`VNew` (parsed version), `State`, `IsLatest` (true when `--pinLatest` detected a bare `latest` tag), `Err` (concrete error when `State == StateError`).
- **State** (`manifest.go`): `todo → check → {skipped | kept | updated | error}`, plus `ignored` for names that don't match any pattern.
- **Pattern**: glob over dependency names; `!` prefix excludes.

## Upgradable vs Skipped Versions

The version regex (`manifest.go`): `^\s*(?:[\^~]\s*)?(\d+[^\s<>|=]*)\s*$`

- **Matches → `StateCheck`**: strings starting with a digit, optionally preceded by `^`/`~`. e.g. `1.2.3`, `^1.2.3`, `~2.3.4`, `1.x`, `1.0.0-beta.1`. The prefix is preserved on update: `^1.2.3` → `^2.0.0`.
- **No match → `StateSkipped`**: comparator ranges (`>=1.0.0`, `<2.0.0`), tags (`latest`), git/file URLs — anything containing `<>|=`.
- The regex is permissive (`1.x` matches), but semver comparison happens later in `engine.go`; invalid semver bypasses the "is it actually newer?" guard.
- **`latestRe`** (`manifest.go`): `(?i)^\s*latest\s*$` — anchored, case-insensitive, matches bare `latest` strings only when `--pinLatest` is active. When `IsLatest=true`, `resolveSpecVersion` sets `SNew = VNew` directly (no regex replacement needed) and `shouldUpdate` short-circuits to always update.

## Gotchas

- **JSON handling**: The package uses Go's `encoding/json/v2` + `encoding/json/jsontext` (requires `GOEXPERIMENT=jsonv2`). `npm.go` uses struct-based unmarshaling for `dist-tags.latest` and `versions` keys. `packagejson.go` uses `jsontext.Decoder` streaming with `InputOffset()` + `ReadValue()` for byte-precise surgical edits in `UpdateDependency`. **Critical**: `jsontext.Token` values are voided by the next decoder call — always call `.String()` before any subsequent `ReadToken`/`ReadValue`/`SkipValue`.
- **Byte-splice safety**: `PackageFile.UpdateDependency` creates a fresh `jsontext.Decoder` on _every_ call, so reported offsets are always current. Successive updates stay correct.
- **`kept` vs `updated`**: if the resolved version is not semver-greater than the current (`engine.go`), the spec becomes `kept`, not `updated`. Exception: `IsLatest` specs always update.
- **Version resolution**: default = `dist-tags.latest`; `-g`/`--greatest` = highest semver across `versions`.
- **Write gate**: the file is rewritten only when `updates > 0 && !cfg.Nop`.
- **Atomic write**: `PackageFile.Write` goes through `github.com/larsartmann/go-atomic-write` (v0.5.1), which stages a temp file with a random suffix, fsyncs it, verifies the on-disk fingerprint still matches the one captured at read time, then performs a single atomic rename and fsyncs the parent directory. This protects against TOCTOU loss when another process (pnpm install, IDE formatter) edits `package.json` during upd's network-fetch window. On mismatch it returns `ErrConcurrentModification` (translated to `upd.ErrConcurrentModification`) and does not touch the file. No `.bak` artifacts are left behind.
- **Quiet path**: `-q`/quiet suppresses the progress bar, table output, AND warnings. The fetch+apply logic is now consolidated — single code path for quiet and non-quiet (reporter is set to noop when quiet).
- **JSON output**: `--json` emits machine-readable JSON to stdout instead of the table. Output includes summary (updated/kept/errors/total), package list, and error details. Intended for CI pipelines.
- **Retry logic**: `npm.go` retries 429/5xx responses with exponential backoff (1s base, 30s cap). `Retry-After` header honored if present. Non-retryable errors (404, network errors) fail immediately. The `RegistryClient.sleep` field (type `sleeper`) abstracts delays so tests run without real sleeps — `newTestEngine` sets it to a no-op; npm_test.go tests can capture delays to assert backoff timing. Retry-error matching uses Go 1.26+ `errors.AsType` (`npm.go:97`, adopted in `b6110bf`).
- **Signal handling**: SIGINT/SIGTERM cancels the fetch phase via `fang.WithNotifySignal` (`cmd/upd/main.go:30`). In-flight HTTP requests are cancelled gracefully.
- **Auto color detection**: `NO_COLOR` env var and non-TTY stdout automatically disable colors for upd's own output. `-C`/`--no-color` and `UPD_NO_COLOR` additionally disable fang's styled help/error colors via a custom `ColorSchemeFunc`. Fang's `colorprofile` writer still handles `NO_COLOR` and non-TTY stdout on its own.
- **Env-var flags**: All public flags can be set via environment variables (`UPD_QUIET`, `UPD_REGISTRY`, `UPD_FILE`, `UPD_TIMEOUT`, etc.). `applyEnvFlags` sets the flag default before Cobra parses CLI args, so explicit flags still override env vars. Invalid env values are ignored and the flag's built-in default is kept. The hidden `--noColor` alias and `--version` do not have env vars.
- **`ProgramVersion`** defaults to `"dev"`; set via `-ldflags -X` at build time. Version is rendered via a custom Cobra template.
- **Error classification** (`npm.go`): `classifyRegistryError` splits HTTP failures into `ErrPackageNotFound` (404/410 — Rejection, exit 1) vs `ErrRegistryUnavailable` (5xx/timeout — Transient, exit 75). This lets CI scripts distinguish retryable from permanent failures. Exit codes derived from `errorfamily.Family.ExitCode()` — Rejection=1, Transient=75, Corruption=65 (EX_DATAERR), Conflict=1, Infrastructure=69.
- **Warnings pipeline** (`cmd/upd/main.go`): `BuildManifest` returns `[]string` warnings for malformed sections and invalid glob patterns. These print as yellow `WARNING:` lines on stderr but don't stop execution. A malformed `upd` field in `package.json` is fatal (stops the run); malformed sections/patterns are non-fatal.
- **Partial failure** (`cmd/upd/main.go`): when `errCount > 0` (one or more packages failed to resolve), `finalizeRun` returns `ErrPartialFailure` after successfully writing the file for packages that did update. Exit code is 1. Successful updates are NOT lost — the file is written before the error is returned.
- **`go-error-family` adoption**: All 13 domain sentinels (`errors.go`) use `errorfamily.NewRejection/NewCorruption/NewTransient/NewConflict` constructors. `ErrHelp`/`ErrVersion` remain plain `errors.New` (control-flow signals, not domain errors). All `fmt.Errorf` wrapping sites in `npm.go`, `packagejson.go`, `engine.go`, `config.go`, `render.go`, `cmd/upd/main.go` use `errorfamily.Wrap*` or `sentinel.WithContext()`. Exit codes derived from `Family.ExitCode()` via `errorfamily.ExitCode(err)`. The `retryableError` wrapper in `npm.go` stays (carries `retryAfter` which errorfamily doesn't model).
- **File-naming trap**: the registry client is `npm.go` (+ `npm_test.go`). Commit `7402f06` renamed _documentation references_ to `pnpm.go` but never renamed the actual file — older status reports may say `pnpm.go`; that file does not exist. Don't "fix" code to match those references.
- **Version truth is split right now**: `flake.nix:22` pins `version = "1.2.0"` while the newest tag is `v1.3.0` (cut 2026-08-16, tag commit `59bcb48`). Binaries built from source therefore report `upd 1.2.0`. Fix tracked as TODO_LIST #1; also note the GitHub Release for v1.3.0 is missing.
- **Legacy `2.x` tags**: `git tag` lists old JS-era tags (`2.9.2`, ... 2020–2023) that sort above the current `v1.x` line in plain output. Use `git tag --sort=-creatordate`. They are not part of this Go release line.
- **Local `go.work` breaks the gates**: an untracked `go.work` (created 2026-09-15) adds `use /home/lars/projects/go-atomic-write`, whose module requires go ≥ 1.27 while the local toolchain is go 1.26.7 with `GOTOOLCHAIN=local` — every plain `go` command in this repo fails until that workspace is removed or the toolchain moves to 1.27. Run the gates with `GOWORK=off` (e.g. `GOEXPERIMENT=jsonv2 GOWORK=off go test ./...`). `nix build`/`nix flake check` are unaffected (filtered source); `nix run .#test`/`.#lint` execute in the working dir and DO hit it.
- **gobwas/glob v1.0.0 API rename**: upstream changed `glob.Glob` (interface, v0.2.x) to `*glob.Pattern` when the v1.0.0 tag was refreshed; the `a98c59a` dependency bump (2026-09-15, post-v1.3.0 auto-commit) shipped without updating `manifest.go`, breaking the build for 10 days (fixed 2026-09-25 — `v1.3.0` itself is unaffected: it still pairs the old API with glob v0.2.3). If other deps get major-bumped by automation, expect API renames to need code changes — don't trust dep-only commits to be compile-clean.
- **Linter: ERRORFAMILY_ADOPT**: Adopted. Only 2 remaining hits (`ErrHelp`, `ErrVersion`) — deliberately kept as `errors.New` because they are control-flow signals (show help, show version), not domain errors.
- **Linter: PHANTOM** (branching-flow): 56 violations for primitive types that "should" be phantom types. **Deliberately not adopted** — over-engineering for a focused CLI. `State` is already a named type; other primitives are clear from context.
- **Linter: STRONG-ID** (branching-flow): 1 violation for `mid string` in `render.go:writeBorder`. **False positive** — `mid` means "middle border position" (top/mid/bottom), not a database ID.
- **Linter: BOOLBLIND** (branching-flow): 1 violation for `Config` struct (8 bool fields). **Deliberately not adopted** — bool config fields is idiomatic Go.
- **Linter: MIXINS** (branching-flow): 1 low-confidence opportunity. **Skipped**.
- **Linter: CONTEXT** (branching-flow): 14 MEDIUM context issues. **Addressed** — errorfamily's `WithContext(key, value)` now attaches structured context at error creation sites. Remaining issues are for complex types (manifest, decoder) that are intentionally suppressed.

## Dependencies (intentional — 8 direct)

- `charm.land/fang/v2` — styled Cobra help, error rendering, man pages, and signal-aware execution.
- `charm.land/lipgloss/v2` — color/style primitives used by fang; used directly to implement the no-color `ColorSchemeFunc`.
- `github.com/spf13/cobra` — command framework, flag parsing, shell-completion, and man-page scaffolding.
- `github.com/spf13/pflag` — POSIX-style short/long flag parsing (used by Cobra).
- `github.com/Masterminds/semver/v3` — semver parse + compare.
- `github.com/gobwas/glob` — dependency-name pattern matching.
- `github.com/larsartmann/go-atomic-write` — TOCTOU-safe atomic file write (fingerprint verify + rename).
- `github.com/larsartmann/go-error-family` — structured error classification (Family, exit codes, context, retry decisions).
- `encoding/json/v2` + `encoding/json/jsontext` — standard library JSON (requires `GOEXPERIMENT=jsonv2`).

## Testing

- Tests are in package `upd` (white-box) alongside source. Shared helpers in `testhelpers_test.go` / `config_test.go`; per-file helpers inline (e.g. `newStatusServer`, `fetchAndApply`, `setupPinLatestTest` in `engine_test.go`, `newCountingServer`/`fetchAndCaptureDelays` in `npm_test.go`, `writeTempPackageJSON` in `integration_test.go`, `newErrorManifest`/`newVerboseErrorManifest` in `render_test.go`, `renderJSONAndParse` in `render_json_test.go`).
- **Zero jscpd clones**: all test duplication eliminated via helper extraction. `jscpd --pattern "**/*.go" --min-lines 5 --min-tokens 40 .` reports 0 clones.
- No network in unit tests — packuments and package files are built from literals.
- Run the full suite before declaring done: `nix run .#test`.
- Race detector is included in CI: `GOEXPERIMENT=jsonv2 go test -race ./...`.
