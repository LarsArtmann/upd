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

- ~~Prebuilt, cross-compiled release binaries (goreleaser or equivalent)~~ shipped 2026-10-05 — `.goreleaser.yml` + tag-triggered `release.yml` (`cfb5cfd`)
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
- Friendlier error when `GOEXPERIMENT=jsonv2` is missing from a plain `go` invocation (today: a cryptic import error)
- Architecture diagram (D2) in `docs/` for onboarding and review contexts
- Config file support (`.updrc`, `upd.json`) beyond the embedded `upd` field
- `--debug` log level / structured logging via `slog`
- Dry-run diff output: show the exact byte diff that would be applied
- Terminal-width-aware error rendering and table layout
- fang error-rendering polish: suppress the auto-appended period when messages already end with punctuation; show positional `[pattern ...]` in the usage line

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
- Range-preserving resolution policies (e.g. bump `^19` within the range instead of rewriting to `^19.x.y`, a `--major-only` mode)
- Other ecosystems (Go modules, Cargo, pip) behind a registry/manifest interface

### 4. Release engineering as a product

Releases run through `.goreleaser.yml` + a tag-triggered workflow, and the
Nix build derives the version from the git revision, so version truth no
longer lives in a hand-edited constant. Direction: a release should be one
auditable, reproducible action — the remaining gap is signing/provenance and
release-day verification.

Raw ideas:

- release-please or an equivalent changelog/tag/bump automator
- Performance benchmark run in CI to catch regressions in diff/glob/manifest hot paths
- Release-smoke job: build on tag push, assert `upd --version` matches the tag
- `nix flake check --all-systems` in the release gate (the default check skips aarch64/darwin)
- golangci-lint SARIF output wired to the GitHub Security tab
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
