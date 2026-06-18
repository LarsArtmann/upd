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
```

Plain Go equivalents: `go build ./cmd/upd`, `go test ./...`, `go vet ./...`.
CI (`.github/workflows/ci.yml`) runs build + vet + test on push/PR to `master`.

## Conventions

- **Nix-first**: all task automation lives in `flake.nix`. Do not add a Makefile or justfile.
- **Strict linting**: `.golangci.yml` enables 100+ linters (errcheck, wrapcheck, varnamelen, exhaustruct, depguard, ...). Expect loud diagnostics on untouched files — match surrounding style rather than chasing every pre-existing warning. `depguard` restricts non-stdlib imports.
- **Single root package**: all library code is package `upd` at the repo root; `cmd/upd` is the only executable.
- **git-town**: `git-town.toml` configures the branch workflow.

## Execution Pipeline

`cmd/upd/main.go:run()` drives a linear flow; the files below each own one stage:

1. **Parse flags** → `Config` (`config.go`). Short and long forms are both registered (`-h`/`--help`, `-c`/`--concurrency`, ...).
2. **Read `package.json`** → `PackageFile` keeps the raw bytes (`packagejson.go`).
3. **Merge patterns**: if `package.json` has an `upd` field (string or array), those args are **prepended** to CLI patterns.
4. **Build manifest** → `Manifest` (`name → []*Spec`) across four sections in fixed order: `optionalDependencies`, `peerDependencies`, `devDependencies`, `dependencies` (`manifest.go`).
5. **Classify** each spec into a `State` via the version regex + glob pattern matching.
6. **Fetch** packuments concurrently (semaphore bounded by `-c`, default 8) from `registry.npmjs.org` (`engine.go`, `npm.go`). Names are lowercased before fetch.
7. **Apply updates**: resolve target version, compare semver, mutate `PackageFile` bytes in place.
8. **Render** terminal table (`render.go`) and **write back** only if updates occurred and `-n` is not set.

## Domain Concepts

- **Packument**: NPM registry JSON for one package (`dist-tags`, `versions`, ...). Held as raw bytes in `npm.go`.
- **Manifest**: `map[string][]*Spec` — every occurrence of a dependency across all sections.
- **Spec**: one dependency in one section. Fields: `Section`, `Name`, `SOld`/`SNew` (full constraint string), `VOld`/`VNew` (parsed version), `State`.
- **State** (`manifest.go`): `todo → check → {skipped | kept | updated | error}`, plus `ignored` for names that don't match any pattern.
- **Pattern**: glob over dependency names; `!` prefix excludes.

## Upgradable vs Skipped Versions

The version regex (`manifest.go`): `^\s*(?:[\^~]\s*)?(\d+[^\s<>|=]*)\s*$`

- **Matches → `StateCheck`**: strings starting with a digit, optionally preceded by `^`/`~`. e.g. `1.2.3`, `^1.2.3`, `~2.3.4`, `1.x`, `1.0.0-beta.1`. The prefix is preserved on update: `^1.2.3` → `^2.0.0`.
- **No match → `StateSkipped`**: comparator ranges (`>=1.0.0`, `<2.0.0`), tags (`latest`), git/file URLs — anything containing `<>|=`.
- The regex is permissive (`1.x` matches), but semver comparison happens later in `engine.go`; invalid semver bypasses the "is it actually newer?" guard.

## Gotchas

- **Byte-splice safety**: `PackageFile.UpdateDependency` re-runs `gjson.GetBytes(p.raw, section)` on _every_ call, so reported offsets are always current. Successive updates stay correct — but never cache gjson results across mutations.
- **`kept` vs `updated`**: if the resolved version is not semver-greater than the current (`engine.go`), the spec becomes `kept`, not `updated`.
- **Version resolution**: default = `dist-tags.latest`; `-g`/`--greatest` = highest semver across `versions`.
- **Write gate**: the file is rewritten only when `updates > 0 && !cfg.Nop`.
- **Quiet path**: `-q`/quiet suppresses the progress bar and takes a separate code branch in `main.go` (fetch without a reporter). The two branches duplicate fetch+apply logic — consolidate carefully if refactoring.
- **`ProgramVersion`** defaults to `"dev"`; set via `-ldflags -X` at build time.

## Dependencies (intentional — only 3 direct)

- `github.com/Masterminds/semver/v3` — semver parse + compare.
- `github.com/gobwas/glob` — dependency-name pattern matching.
- `github.com/tidwall/gjson` — JSON path reads + byte offsets for surgical edits.

## Testing

- Tests are in package `upd` (white-box) alongside source. Helpers in `testhelpers_test.go` / `config_test.go`.
- No network in unit tests — packuments and package files are built from literals.
- Run the full suite before declaring done: `nix run .#test`.
