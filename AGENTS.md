# UPD — Agent Context

Concise, enduring context for AI sessions working on `upd`.

## What This Is

`upd` is a Go CLI that upgrades JavaScript package dependencies in an NPM `package.json` while **strictly preserving the original JSON formatting and whitespace**. It intentionally skips version constraint formulas (`^`, `~`, ranges) and only bumps the concrete version embedded in the constraint string.

Originally written in JavaScript by Dr. Ralf S. Engelschall; ported to Go by Lars Artmann. The Go port keeps the same CLI behavior while using Go's performance, type safety, and single-binary distribution.

## Build, Test, Lint

Use Nix flakes (preferred in this repo):

```bash
nix run .#build   # go build ./cmd/upd -> bin/upd
nix run .#test    # go test ./... -v -count=1
nix run .#lint    # go vet ./... && go build ./...
nix run .#run -- <args>   # go run ./cmd/upd <args>
```

Plain Go equivalents:

```bash
go build ./cmd/upd
go test ./...
go vet ./...
```

CI runs `go build`, `go vet`, and `go test ./... -v -count=1` on pushes/PRs to `master`.

## Architecture

Single Go module (`github.com/LarsArtmann/upd`) with a small library package and a thin `cmd/upd` main.

```
cmd/upd/main.go     # CLI entry point: parse flags, read package.json,
                    # build manifest, fetch registry, apply updates, render, write.
config.go           # Config, flag parsing, usage/version text.
engine.go           # Engine orchestrates registry fetches and applies updates.
npm.go              # RegistryClient fetches packuments from registry.npmjs.org.
packagejson.go      # PackageFile reads/updates package.json preserving formatting.
manifest.go         # Manifest + Spec model; pattern matching against dependency names.
render.go           # Terminal table rendering with ANSI colors/diff highlighting.
progress.go         # Progress reporter used during registry fetches.
diff.go             # Character diff for highlighting version changes.
errors.go           # Sentinel errors.
```

## Domain Concepts

- **Packument**: NPM registry JSON document for a single package (contains `dist-tags`, `versions`, etc.).
- **Manifest**: In-memory map of dependency names to `[]*Spec`, built from `package.json` sections and filtered by glob patterns.
- **Spec**: One dependency occurrence in one section. Tracks section, name, old/new constraint strings (`SOld`/`SNew`), old/new parsed versions (`VOld`/`VNew`), and `State`.
- **State**: Lifecycle of a dependency entry — `todo`, `check`, `skipped`, `kept`, `updated`, `error`, `ignored`.
- **Pattern**: Glob pattern matched against dependency names. Prefix with `!` to exclude.

See `docs/DOMAIN_LANGUAGE.md` for the formal glossary (currently a placeholder; update when terminology stabilizes).

## Conventions

- **Package name**: `upd` (root package contains all library code). `cmd/upd` is the executable.
- **Formatting preservation is critical**: `PackageFile.UpdateDependency` replaces only the raw JSON bytes of the changed value; it does not re-encode the whole file.
- **Registry URLs**: hardcoded to `https://registry.npmjs.org`. Package names are lowercased before fetching.
- **Versions**: resolved via `dist-tags.latest` (default) or greatest semver version (with `-g`/`--greatest`).
- **Dependency sections** are checked in this order: `optionalDependencies`, `peerDependencies`, `devDependencies`, `dependencies`.
- **Embedded config**: `package.json` may contain an `upd` field (string or array of strings) that is prepended to CLI patterns.
- **Errors**: sentinel errors live in `errors.go` and are wrapped with context at the call site.

## Dependencies

Only three direct dependencies, all intentional:

- `github.com/Masterminds/semver/v3` — semver parsing and comparison.
- `github.com/gobwas/glob` — glob pattern matching for dependency names.
- `github.com/tidwall/gjson` — fast JSON path reads and validation; also used to locate byte offsets for surgical replacements.

## Gotchas

- `PackageFile.UpdateDependency` mutates `p.raw` in place by byte offset. If multiple specs for the same dependency exist in different sections, earlier replacements shift later offsets. The code currently updates each section independently; be very careful if changing this.
- Constraints like `^1.2.3` are parsed to extract `1.2.3`, then the new version replaces only the version portion: `^1.2.3` → `^2.0.0`. The prefix is preserved.
- If the new resolved version is not semver-greater than the old one, the spec is marked `kept` rather than `updated`.
- `FetchAll` uses a semaphore bounded by `Config.Concurrency` (default 8).
- CLI short flags use single-letter forms (`-h`, `-V`, `-q`, `-n`, `-C`, `-f`, `-g`, `-a`, `-c`). Long forms are also accepted.
- `ProgramVersion` is set via `-ldflags` at build time; default value is `"dev"`.

## Testing Patterns

- Tests live in `*_test.go` files in the root package (the same package as the code under test).
- Tests use plain `testing` and small project-specific helpers in `testhelpers_test.go` and `config_test.go`.
- Network-dependent paths are avoided in unit tests; test packuments and package files are built from literals or testdata when possible.
