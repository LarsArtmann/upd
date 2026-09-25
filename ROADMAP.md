# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to `TODO_LIST.md`.

## Themes

### 1. Frictionless distribution

`upd` today requires either Nix or a Go 1.26+ toolchain with
`GOEXPERIMENT=jsonv2`. That is a real adoption barrier for the casual
"just update my deps" user. Direction: anyone should be one command away
from a working `upd` on any platform.

Raw ideas:

- Prebuilt, cross-compiled release binaries (goreleaser or equivalent)
- Homebrew tap and a published Nix flake for `nix profile install`
- Release artifacts with checksums and signatures (cosign/GPG) beyond the signed git tag
- Reproducible-build badge — the Nix build already is reproducible; surface it
- SBOM / dependency manifest attached to each release
- SLSA provenance for release artifacts (advanced)
- Templated release notes generated from `CHANGELOG.md`; changelog lint in CI
- `nix run github:LarsArtmann/upd/<tag>` — verify and document running a specific tag directly

### 2. Trustworthy, observable CLI

The tool mutates a file developers care about; every interaction should
inspire confidence and every failure should explain itself.

Raw ideas:

- `doctor` subcommand: registry reachability + `package.json` sanity check
- `check` subcommand: validate and report only, never write
- Config file support (`.updrc`, `upd.json`) beyond the embedded `upd` field
- `--debug` log level / structured logging via `slog`
- Dry-run diff output: show the exact byte diff that would be applied
- Terminal-width-aware error rendering and table layout

### 3. Beyond a single `package.json`

The architecture (manifest builder → glob filter → semver compare →
byte-splice writer) is general. The open product question, posed since
the first Go-port report and still unanswered: faithful NPM-only tool,
or general dependency updater?

Raw ideas:

- Monorepo workspace support (npm/yarn/pnpm `workspaces`)
- Lockfile awareness (`package-lock.json`, `pnpm-lock.yaml`, `yarn.lock`)
- Batch mode across multiple `package.json` files
- Offline mode with a packument cache for air-gapped use
- Other ecosystems (Go modules, Cargo, pip) behind a registry/manifest interface

### 4. Release engineering as a product

Releases are currently manual and version truth lives only in
`flake.nix`. Direction: a release should be one auditable, reproducible
action.

Raw ideas:

- release-please or an equivalent changelog/tag/bump automator
- CI build injecting the git-describe version instead of the literal `"ci"` string
- Performance benchmark run in CI to catch regressions in diff/glob/manifest hot paths
- Release-smoke job: build on tag push, assert `upd --version` matches the tag
- A documented deprecation policy (relevant once breaking changes are considered)
- An announce-channel note (GitHub Releases only, or more)

## Non-goals

Things we are deliberately NOT pursuing and why:

- **Docker images:** a single static binary makes containers pointless overhead.
- **Self-update command / update notifier:** Go binaries don't self-update;
  installation is owned by nix / `go install` / future package managers.
- **Structured-error HTTP/daemon mode:** `upd` is a short-lived CLI; no server
  surface is planned.
- **Faithful parity with `rse/upd` internals:** this is a rewrite, not a port
  of implementation details; behavioral parity only where it serves users.

---

_Created 2026-09-25 from the docs-health audit of `docs/status/` reports and
`docs/research/` notes. Revisit quarterly; graduate items to `TODO_LIST.md`
when they become bounded._
