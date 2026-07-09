# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

- `--pinLatest` / `-P` flag to pin dependencies using the bare `latest` dist-tag
  to their exact resolved semver version (e.g. `"latest"` → `"7.7.4"`)
- `encoding/json/v2` + `encoding/json/jsontext` migration, replacing `tidwall/gjson`
- VHS animated demos rendered to `vhs.charm.sh` cloud

### Changed

- Copyright year updated to 2026
- Fixed swapped VERSION OLD / VERSION NEW columns in noColor (`-C`) mode
  — the diff-highlight fallback now returns the correct string for each column

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
