# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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
