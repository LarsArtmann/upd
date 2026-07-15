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
nix run .#lint           # go vet ./... && go build ./...
nix run .#run -- <args>  # go run ./cmd/upd <args>
nix run .#demo           # render VHS demo GIFs locally
nix run .#demo -- --publish  # render + publish to vhs.charm.sh
```

**`GOEXPERIMENT=jsonv2` is required** for all Go commands (set automatically in nix apps and devShell; must be `export`ed for plain Go commands).

Plain Go equivalents: `GOEXPERIMENT=jsonv2 go build ./cmd/upd`, `GOEXPERIMENT=jsonv2 go test ./...`.
CI (`.github/workflows/ci.yml`) runs build + vet + race test on push/PR to `master`.

VHS demos (`demo/*.tape`) are rendered with `vhs` and published to the [VHS cloud](https://vhs.charm.sh). GIFs are git-ignored — re-render with `nix run .#demo` anytime without committing binary files.

## Conventions

- **Nix-first**: all task automation lives in `flake.nix`. Do not add a Makefile or justfile.
- **Strict linting**: `.golangci.yml` enables 100+ linters (errcheck, wrapcheck, varnamelen, exhaustruct, depguard, ...). Expect loud diagnostics on untouched files — match surrounding style rather than chasing every pre-existing warning. `depguard` restricts non-stdlib imports.
- **Single root package**: all library code is package `upd` at the repo root; `cmd/upd` is the only executable.
- **git-town**: `git-town.toml` configures the branch workflow.

## Execution Pipeline

`cmd/upd/main.go:run()` drives a linear flow; the files below each own one stage:

1. **Parse flags** → `Config` (`config.go`). Short and long forms are both registered (`-h`/`--help`, `-c`/`--concurrency`, `-P`/`--pin-latest`, `-r`/`--registry`, `-t`/`--timeout`, `--retries`, `--json`, `--verbose`, `--dry-run`). Auto color detection via `ShouldDisableColor` (`NO_COLOR` env var, non-TTY stdout).
2. **Read `package.json`** → `PackageFile` keeps the raw bytes and an xxhash64 fingerprint of them (`packagejson.go`). The fingerprint guards the later write against concurrent modifications.
3. **Merge patterns**: if `package.json` has an `upd` field (string or array), those args are **prepended** to CLI patterns.
4. **Build manifest** → `(Manifest, []string)` — the manifest is `name → []*Spec` across four sections in fixed order: `optionalDependencies`, `peerDependencies`, `devDependencies`, `dependencies` (`manifest.go`). The second return value is a slice of warning strings (malformed sections, invalid glob patterns). Takes `pinLatest bool` as third arg. `GetUpdArgs()` and `GetDependencySection()` return `error` as second value.
5. **Classify** each spec into a `State` via the version regex (`versionRe`) + `latestRe` + glob pattern matching.
6. **Fetch** packuments concurrently (semaphore bounded by `-c`, default 8) from registry (`engine.go`, `npm.go`). Names are lowercased before fetch. `RegistryClient` takes `*Config` (supports custom registry URL, timeout, retries). Retry logic: 429/5xx retried with exponential backoff + `Retry-After` header support. HTTP transport tuned (MaxIdleConns, IdleConnTimeout). Context is signal-aware (SIGINT/SIGTERM cancellation).
7. **Apply updates**: resolve target version, compare semver, mutate `PackageFile` bytes in place. When `IsLatest=true`, `SNew` is set directly. Errored specs carry their concrete error in `Spec.Err`.
8. **Render** terminal table (`render.go`) or JSON output (`--json`, `RenderJSON`). Error detail block supports `--verbose`. **Write back** only if updates occurred and `-n` is not set. Atomic write via `go-atomic-write`.
9. **Exit code** (`cmd/upd/main.go`): `ErrRegistryUnavailable` → exit 75; `ErrPartialFailure` → exit 1; all other errors → exit 1; success → exit 0. Warnings print to stderr but are suppressed in quiet mode.

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
- **Atomic write**: `PackageFile.Write` goes through `github.com/larsartmann/go-atomic-write` (v0.2.0+), which stages a temp file with a random suffix, fsyncs it, verifies the on-disk fingerprint still matches the one captured at read time, then performs a single atomic rename and fsyncs the parent directory. This protects against TOCTOU loss when another process (npm install, IDE formatter) edits `package.json` during upd's network-fetch window. On mismatch it returns `ErrConcurrentModification` (translated to `upd.ErrConcurrentModification`) and does not touch the file. No `.bak` artifacts are left behind.
- **Quiet path**: `-q`/quiet suppresses the progress bar, table output, AND warnings. The fetch+apply logic is now consolidated — single code path for quiet and non-quiet (reporter is set to noop when quiet).
- **JSON output**: `--json` emits machine-readable JSON to stdout instead of the table. Output includes summary (updated/kept/errors/total), package list, and error details. Intended for CI pipelines.
- **Retry logic**: `npm.go` retries 429/5xx responses with exponential backoff (1s base, 30s cap). `Retry-After` header honored if present. Non-retryable errors (404, network errors) fail immediately. The `RegistryClient.sleep` field (type `sleeper`) abstracts delays so tests run without real sleeps — `newTestEngine` sets it to a no-op; npm_test.go tests can capture delays to assert backoff timing.
- **Signal handling**: SIGINT/SIGTERM cancels the fetch phase via `signal.NotifyContext`. In-flight HTTP requests are cancelled gracefully.
- **Auto color detection**: `NO_COLOR` env var and non-TTY stdout automatically disable colors. `-C` is still available for manual override.
- **`ProgramVersion`** defaults to `"dev"`; set via `-ldflags -X` at build time.
- **Error classification** (`npm.go`): `classifyRegistryError` splits HTTP failures into `ErrPackageNotFound` (404/410 — user typo, exit 1) vs `ErrRegistryUnavailable` (5xx/timeout — system fault, exit 75). This lets CI scripts distinguish retryable from permanent failures.
- **Warnings pipeline** (`cmd/upd/main.go`): `BuildManifest` returns `[]string` warnings for malformed sections and invalid glob patterns. These print as yellow `WARNING:` lines on stderr but don't stop execution. A malformed `upd` field in `package.json` is fatal (stops the run); malformed sections/patterns are non-fatal.
- **Partial failure** (`cmd/upd/main.go`): when `errCount > 0` (one or more packages failed to resolve), `finalizeRun` returns `ErrPartialFailure` after successfully writing the file for packages that did update. Exit code is 1. Successful updates are NOT lost — the file is written before the error is returned.
- **Linter: ERRORFAMILY_ADOPT**: The `branching-flow` linter suggests replacing `errors.New`/`fmt.Errorf` with "go-error-family" constructors (`NewRejection`, `WrapTransient`). This is **deliberately not adopted** — the project intentionally keeps only 3 direct dependencies, `depguard` restricts non-stdlib imports, and the current sentinel-error + `fmt.Errorf("context: %w", err)` pattern is idiomatic Go. Do not add a 4th dependency to satisfy this rule.
- **Renderer API**: `NewRenderer(w io.Writer, opts RendererOptions)` takes an options struct, not positional bools. `RendererOptions{NoColor, Verbose bool}`. This prevents `NewRenderer(w, true, false)` ambiguity.
- **Linter: PHANTOM** (branching-flow): 56 violations for primitive types (string/int/bool params and fields) that "should" be phantom types. **Deliberately not adopted** — wrapping every `string`/`int` in a named type would be massive over-engineering for a focused CLI. `State` is already a named type (7 distinct domain values). Other primitives (`short`, `long`, `section`, `updates`, etc.) are clear from context and don't benefit from phantom wrapping.
- **Linter: STRONG-ID** (branching-flow): 1 violation for `mid string` in `render.go:writeBorder`. **False positive** — `mid` means "middle border position" (top/mid/bottom), not a database ID. The name coincidentally contains "id".
- **Linter: BOOLBLIND** (branching-flow): 1 violation for `Config` struct (8 bool fields → bit flags). **Deliberately not adopted** — bool fields on a config struct is idiomatic Go. Bit flags (`cfg.Flags&FlagGreatest != 0`) add ceremony without readability gain. The 7-byte savings (8B→1B) is irrelevant for a config struct created once.
- **Linter: MIXINS** (branching-flow): 1 low-confidence opportunity for `FetchResult`/`entry` shared fields. **Skipped** — only 2 structs share fields, and they serve different purposes (fetch result vs. manifest entry).
- **Linter: CONTEXT** (branching-flow): 14 MEDIUM error-context issues. **Skipped** — the caller (`engine.applyOne`) already wraps errors with `name` and `section` context. Adding redundant context at every level creates noise. Current pattern: errors include context where first available; callers add broader context. This is the idiomatic Go error-wrapping pattern.
- **`usageBlankLine`**: Previously had an infinite recursion bug (`usageBlankLine` called itself instead of `fmt.Fprintln(w)`). Fixed — SA5007 (staticcheck) would catch this.

## Dependencies (intentional — only 3 direct)

- `github.com/Masterminds/semver/v3` — semver parse + compare.
- `github.com/gobwas/glob` — dependency-name pattern matching.
- `github.com/larsartmann/go-atomic-write` — TOCTOU-safe atomic file write (fingerprint verify + rename).
- `encoding/json/v2` + `encoding/json/jsontext` — standard library JSON (requires `GOEXPERIMENT=jsonv2`).

## Testing

- Tests are in package `upd` (white-box) alongside source. Shared helpers in `testhelpers_test.go` / `config_test.go`; per-file helpers inline (e.g. `newStatusServer`, `fetchAndApply`, `setupPinLatestTest` in `engine_test.go`, `newCountingServer`/`fetchAndCaptureDelays` in `npm_test.go`, `writeTempPackageJSON` in `integration_test.go`, `newErrorManifest`/`newVerboseErrorManifest` in `render_test.go`, `renderJSONAndParse` in `render_json_test.go`).
- **Zero jscpd clones**: all test duplication eliminated via helper extraction. `jscpd --pattern "**/*.go" --min-lines 5 --min-tokens 40 .` reports 0 clones.
- No network in unit tests — packuments and package files are built from literals.
- Run the full suite before declaring done: `nix run .#test`.
- Race detector is included in CI: `GOEXPERIMENT=jsonv2 go test -race ./...`.
