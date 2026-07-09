# UPD

**Upgrade NPM Package Dependencies**

A fast Go CLI for upgrading the JavaScript package dependencies in a
Node Package Manager (NPM) `package.json` configuration file while
strictly preserving the formatting of the existing JSON syntax and
intentionally skipping version constraint formulas.

> **Go rewrite of [`rse/upd`](https://github.com/rse/upd)** — the
> original JavaScript/Node.js CLI by
> [Dr. Ralf S. Engelschall](https://engelschall.com/).

## Origin

- **Original:** [`rse/upd`](https://github.com/rse/upd) — an
  [npm](https://www.npmjs.com/package/upd) package written in
  JavaScript/Node.js by
  [Dr. Ralf S. Engelschall](https://engelschall.com/).
- **This project:** a complete [Go](https://go.dev/) rewrite by
  [Lars Artmann](https://lars.software/), keeping the same CLI behavior
  and philosophy while leveraging Go's performance, compile-time type
  safety, and single-binary distribution.

## Installation

```
$ go install github.com/LarsArtmann/upd/cmd/upd@latest
```

## Usage

```
$ upd [-h] [-V] [-q] [-n] [-C] [-f <file>] [-g] [-a] [-c <concurrency>] [-P] [<pattern> ...]
```

- `-h`, `--help`<br/>
  Show usage help.
- `-V`, `--version`<br/>
  Show program version information.
- `-q`, `--quiet`<br/>
  Quiet operation (do not output upgrade information).
- `-n`, `--nop`<br/>
  No operation (do not modify package configuration file).
- `-C`, `--noColor`<br/>
  Do not use any colors in output.
- `-f <file>`, `--file <file>`<br/>
  Package configuration to use ("package.json").
- `-g`, `--greatest`<br/>
  Use greatest version (instead of latest stable one).
- `-a`, `--all`<br/>
  Show all packages (instead of just updated ones).
- `-c <concurrency>`, `--concurrency <concurrency>`<br/>
  Number of concurrent network connections to NPM registry.
- `-P`, `--pin-latest`<br/>
  Pin "latest" tag to exact version (e.g., `latest` → `1.0.0`).
- `<pattern>`<br/>
  Positive or negative (if prefixed with `!`) Glob pattern for matching
  names of dependencies to update.

## Demo

Animated demos are rendered with [VHS](https://github.com/charmbracelet/vhs)
and published to the [VHS cloud](https://vhs.charm.sh).

[![Made with VHS](https://stuff.charm.sh/vhs/badge.svg)](https://github.com/charmbracelet/vhs)

**Render locally:**

```
$ nix run .#demo
```

**Publish to the cloud (returns a shareable `vhs.charm.sh` URL):**

```
$ nix run .#demo -- --publish
```

Tape sources live in [`demo/`](demo/) and are the single source of truth —
re-render at any time without touching committed GIFs.

## Development

```
$ go build ./cmd/upd           # Build
$ go test ./...                # Test
$ go vet ./...                 # Vet
```

## License

Copyright &copy; 2015-2026 Dr. Ralf S. Engelschall (http://engelschall.com/)

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
