# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

This repo standardizes on [Nix flakes](https://nix.dev/concepts/flakes) for all
build automation — no Makefile, no justfile.

```bash
nix run .#build             # build to bin/upd
nix run .#test              # go test ./... -v -count=1 -race
nix run .#test-integration  # opt-in tests against the real NPM registry
nix run .#lint              # go vet + go build + golangci-lint
nix run .#run -- <args>     # go run ./cmd/upd <args>
```

Run `nix run .#test` and `nix run .#lint` before submitting a PR — CI runs
the same gates on every push.

### Plain Go equivalents

The project uses `encoding/json/v2`, which requires `GOEXPERIMENT=jsonv2`:

```bash
export GOEXPERIMENT=jsonv2   # required for all Go commands
go build ./cmd/upd
go test -race ./...
go vet ./...
```

### Linting

`nix run .#lint` runs `go vet`, a build check, and the full golangci-lint
suite (`.golangci.yml`, 100+ linters). Expect loud diagnostics on first run —
match surrounding style rather than chasing every pre-existing warning.

```bash
golangci-lint run ./...     # full linter suite (optional, strict)
```

## Reporting Issues

Please use [GitHub Issues](https://github.com/LarsArtmann/upd/issues) to report
bugs or request features.
