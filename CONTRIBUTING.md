# Contributing

Thanks for your interest in contributing!

## How to Contribute

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

This repo standardizes on [Nix flakes](https://nixos.wiki/wiki/Flakes) for all
build automation — no Makefile, no justfile.

```bash
nix run .#build          # build to bin/upd
nix run .#test           # go test ./... -v -count=1
nix run .#lint           # go vet ./... && go build ./...
nix run .#run -- <args>  # go run ./cmd/upd <args>
```

### Plain Go equivalents

The project uses `encoding/json/v2`, which requires `GOEXPERIMENT=jsonv2`:

```bash
export GOEXPERIMENT=jsonv2   # required for all Go commands
go build ./cmd/upd
go test -race ./...
go vet ./...
```

### Linting

The project has `.golangci.yml` with 100+ linters enabled. Expect loud
diagnostics on first run — match surrounding style rather than chasing every
pre-existing warning.

```bash
golangci-lint run ./...     # full linter suite (optional, strict)
```

## Reporting Issues

Please use [GitHub Issues](https://github.com/LarsArtmann/upd/issues) to report
bugs or request features.
