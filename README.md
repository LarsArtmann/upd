# UPD

**Upgrade NPM Package Dependencies — fast, safe, formatting-preserving.**

[![CI](https://github.com/LarsArtmann/upd/actions/workflows/ci.yml/badge.svg)](https://github.com/LarsArtmann/upd/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LarsArtmann/upd)](https://goreportcard.com/report/github.com/LarsArtmann/upd)
[![Go Reference](https://pkg.go.dev/badge/github.com/LarsArtmann/upd.svg)](https://pkg.go.dev/github.com/LarsArtmann/upd)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A Go CLI that bumps dependency versions in an NPM `package.json` while
**byte-preserving all original JSON formatting** — whitespace, key order,
and quoting style stay exactly as you wrote them. Only the version
number inside each constraint string changes. Nothing else is touched.

> Go rewrite of [`rse/upd`](https://github.com/rse/upd) — the original
> JavaScript/Node.js CLI by
> [Dr. Ralf S. Engelschall](https://engelschall.com/).

---

## Demo

![upd demo](https://vhs.charm.sh/vhs-44iono1WluRIVsM2ddBpRc.gif)

[![Made with VHS](https://stuff.charm.sh/vhs/badge.svg)](https://github.com/charmbracelet/vhs)

## Features

- **Formatting-preserving edits** — surgically replaces only the version
  bytes inside each constraint string. Your indentation, key order, and
  quoting style are never touched.
- **TOCTOU-safe atomic writes** — stages a temp file, fsyncs it, verifies
  the on-disk fingerprint hasn't changed since read, then atomically renames.
  If another process (npm install, IDE formatter) edited `package.json`
  during the network-fetch window, the write is aborted and your file is
  left untouched.
- **Concurrent registry queries** — fetches packuments in parallel with a
  configurable connection pool (default 8).
- **Semantic version resolution** — resolves to `dist-tags.latest` by
  default, or the highest semver across all published versions with `-g`.
- **Pin `latest` tags** — `-P` / `--pin-latest` rewrites bare `"latest"`
  tags to their exact resolved semver (e.g. `"latest"` → `"7.7.4"`).
- **Glob pattern filtering** — update only matching dependencies, with
  `!`-prefixed exclusions (e.g. `upd react* !react-dom`).
- **All four dependency sections** — `dependencies`, `devDependencies`,
  `peerDependencies`, `optionalDependencies`.
- **Character-level diff highlighting** — shows exactly which characters
  changed in red/green.
- **Embedded defaults** — an `"upd"` field in `package.json` supplies
  default CLI arguments so you don't repeat yourself.
- **Single static binary** — no runtime dependencies, no Node.js required.

## Quick Start

```bash
# Install (requires Go 1.26+ with GOEXPERIMENT=jsonv2)
GOEXPERIMENT=jsonv2 go install github.com/LarsArtmann/upd/cmd/upd@latest

# Dry run — show what would change without writing
upd -n

# Apply updates
upd
```

> **Why `GOEXPERIMENT=jsonv2`?** `upd` uses Go's `encoding/json/v2` for
> byte-precise JSON editing. This flag will become unnecessary once the
> Go team stabilizes the `json/v2` package in a future release.

Example output (`upd -n -C`):

```
┌─────────────────────────────────────┬──────────────┬──────────────┬─────────┐
│MODULE NAME                          │VERSION OLD   │VERSION NEW   │STATE    │
├─────────────────────────────────────┼──────────────┼──────────────┼─────────┤
│express                              │^4.18.0       │^5.2.1        │updated  │
│jest                                 │^29.5.0       │^30.4.2       │updated  │
│lodash                               │^4.17.20      │^4.18.1       │updated  │
│typescript                           │^5.2.0        │^7.0.2        │updated  │
└─────────────────────────────────────┴──────────────┴──────────────┴─────────┘
```

## Installation

### Go

```bash
GOEXPERIMENT=jsonv2 go install github.com/LarsArtmann/upd/cmd/upd@latest
```

Requires Go 1.26+ with the `json/v2` experiment enabled.

### Nix

```bash
# Run directly without installing
nix run github:LarsArtmann/upd

# Install to your profile
nix profile install github:LarsArtmann/upd
```

### Build from source

```bash
git clone https://github.com/LarsArtmann/upd.git
cd upd
nix run .#build    # or: GOEXPERIMENT=jsonv2 go build -o upd ./cmd/upd
```

## Usage

```
upd [-h] [-V] [-q] [-n|--dry-run] [-C] [-f <file>] [-r <registry>] [-g] [-a] [-c <concurrency>] [-P] [-t <timeout>] [--retries <n>] [--json] [--verbose] [<pattern> ...]
```

| Flag | Long form       | Description                                               |
| ---- | --------------- | --------------------------------------------------------- |
| `-h` | `--help`        | Show usage help.                                          |
| `-V` | `--version`     | Show program version.                                     |
| `-q` | `--quiet`       | Suppress output (no progress bar, no table, no warnings). |
| `-n` | `--nop`         | Dry run — do not modify `package.json`.                   |
|      | `--dry-run`     | Alias for `--nop`.                                        |
| `-C` | `--noColor`     | Disable ANSI colors in output.                            |
| `-f` | `--file`        | Path to package config (default: `package.json`).         |
| `-r` | `--registry`    | NPM registry base URL (default: `registry.npmjs.org`).    |
| `-g` | `--greatest`    | Use greatest published version instead of `latest` tag.   |
| `-a` | `--all`         | Show all packages, not just updated ones.                 |
| `-c` | `--concurrency` | Concurrent NPM registry connections (default: 8).         |
| `-P` | `--pin-latest`  | Pin bare `latest` tags to exact semver.                   |
| `-t` | `--timeout`     | Per-request timeout (default: `20s`).                     |
|      | `--retries`     | Max retries for transient 429/5xx failures (default: 3).  |
|      | `--json`        | Machine-readable JSON output for CI/scripts.              |
|      | `--verbose`     | Show full error chains in the error detail block.         |
|      | `<pattern>`     | Glob pattern for dependency names. `!` prefix excludes.   |

**Color auto-detection:** Colors are automatically disabled when the `NO_COLOR`
environment variable is set (see [no-color.org](https://no-color.org/)) or when
stdout is not a terminal (piped/redirected). Use `-C` to force-disable.

**Examples:**

```bash
upd                          # update all dependencies
upd -n                       # dry run (preview changes)
upd react*                   # only update packages matching "react*"
upd react* !react-dom        # update react packages except react-dom
upd -g                       # use greatest version (include pre-releases)
upd -P                       # pin "latest" tags to exact versions
upd -c 16 lodash*            # 16 concurrent connections, only lodash
upd -f alt.json -n           # preview changes to alt.json
```

## How It Works

### Byte-preserving edits

`upd` parses `package.json` with a streaming JSON decoder that tracks
byte offsets. When a version needs updating, it splices only the version
bytes out of the raw file and inserts the new ones. The rest of the file
— every space, newline, key ordering, and quoting choice — is preserved
exactly. No re-serialization, no formatter arguments, no diffs to resolve.

### Version resolution

Each dependency constraint is classified by a regex:

- **Upgradable** — strings starting with a digit, optionally preceded by
  `^` or `~` (e.g. `1.2.3`, `^1.2.3`, `~2.3.4`, `1.x`). The prefix is
  preserved on update: `^1.2.3` → `^2.0.0`.
- **Skipped** — comparator ranges (`>=1.0.0`), tags (`latest`), git/file
  URLs — anything containing `<>|=`. Use `-P` to opt-in to pinning `latest`.
- **Ignored** — names that don't match any supplied glob pattern.

After resolving the target version from the NPM registry, a semver
comparison guards against downgrades: if the resolved version isn't
actually newer, the dependency is marked `kept`.

### Atomic writes with TOCTOU protection

The write path uses [go-atomic-write](https://github.com/larsartmann/go-atomic-write)
to prevent data loss when another process edits `package.json` during the
network-fetch window. It stages a temp file, fsyncs it, acquires a
cross-platform file lock, verifies the on-disk fingerprint hasn't changed
since read, then atomically renames. On mismatch it aborts with
`ErrConcurrentModification` — the file is untouched.

See **[docs/atomic-writes.md](docs/atomic-writes.md)** for the full
step-by-step breakdown and flow diagram.

## Configuration

### Embedded `upd` field

Add an `"upd"` field to your `package.json` to supply default arguments
that are **prepended** to CLI flags:

```json
{
  "upd": ["react*", "!react-dom", "-c", "16"],
  "dependencies": {
    "react": "^18.0.0",
    "react-dom": "^18.0.0",
    "lodash": "^4.17.20"
  }
}
```

Now `upd` is equivalent to `upd react* !react-dom -c 16`. CLI flags
override or supplement these defaults. The field accepts a string or an
array.

## Exit Codes

| Code | Meaning                                                                                      |
| ---- | -------------------------------------------------------------------------------------------- |
| `0`  | Success — all dependencies resolved without errors.                                          |
| `1`  | Failure — package not found, partial resolution errors, IO error, or malformed input.        |
| `75` | Registry unavailable — transient 5xx/timeout from the NPM registry. Retryable (EX_TEMPFAIL). |

In CI, check for exit 75 to decide whether to retry the job. Exit 1 means
something needs human attention (typo in dependency name, malformed JSON, etc.).

## Troubleshooting

**`ERROR: package not found in NPM registry`** (exit 1)
The package name in `package.json` doesn't exist on the registry. Check for
typos, check if the package was unpublished, or verify you're using the correct
registry (`-r`/`--registry`).

**`ERROR: NPM registry is unavailable`** (exit 75)
The registry returned a server error (5xx) or timed out. This is transient —
re-run `upd` after a few seconds. If using a private registry, verify it's
running and accessible. Use `--retries` to increase the number of retry attempts.

**`ERROR: package configuration file was modified concurrently`**
Another process (npm install, IDE auto-save, formatter) edited `package.json`
while `upd` was fetching versions. Your file was not changed. Simply re-run `upd`.

**`ERROR: invalid JSON in package configuration file`**
Your `package.json` has malformed JSON. Run `npx jsonlint package.json` or
`node -e "JSON.parse(require('fs').readFileSync('package.json','utf8'))"` to
find the syntax error.

**Progress bar is garbled or leaves artifacts**
The progress bar is cleared using a fixed-width reset. If your terminal is
narrower than 80 characters, use `-q` (quiet mode) to suppress it.

**Colors appear in piped output**
Colors are auto-disabled when stdout is not a terminal or `NO_COLOR` is set.
If you still see colors, ensure no tool in your pipeline (e.g. `script`) is
emulating a TTY. You can always force-disable with `-C`.

## Development

This repo uses [Nix flakes](https://nixos.wiki/wiki/Flakes) for all
build automation:

```bash
nix run .#build          # build to bin/upd
nix run .#test           # go test ./... -v -count=1
nix run .#lint           # go vet + go build + golangci-lint
nix run .#run -- <args>  # go run ./cmd/upd <args>
nix flake check          # validate the flake
```

Plain Go equivalents (requires `GOEXPERIMENT=jsonv2`):

```bash
export GOEXPERIMENT=jsonv2   # required — uses encoding/json/v2
go build ./cmd/upd
go test -race ./...
go vet ./...
```

### Render the demo

```bash
nix run .#demo              # render GIF locally to demo/
nix run .#demo -- --publish # render + upload to vhs.charm.sh cloud
```

## Origin

- **Original:** [`rse/upd`](https://github.com/rse/upd) — an
  [npm](https://www.npmjs.com/package/upd) package written in
  JavaScript/Node.js by
  [Dr. Ralf S. Engelschall](https://engelschall.com/).
- **This project:** a complete [Go](https://go.dev/) rewrite by
  [Lars Artmann](https://lars.software/), keeping the same CLI behavior
  and philosophy while leveraging Go's performance, compile-time type
  safety, and single-binary distribution.

## License

MIT — Copyright &copy; 2015-2026 Dr. Ralf S. Engelschall,
Copyright &copy; 2026 Lars Artmann.

See [LICENSE](LICENSE) for the full text.
