# Releasing upd

The complete ritual for cutting a release. Every release is `vX.Y.Z` on the
`master` branch — there are no release branches.

## Versioning

- **Patch** (`v1.4.1`): bug fixes only.
- **Minor** (`v1.5.0`): new features, flags, or output fields (additive).
- **Major** (`v2.0.0`): breaking changes — deprecated aliases (`--noColor`,
  `--json`, `UPD_JSON`) are removed here, not before.

The Nix build derives the version string from the git revision
(`flake.nix` uses `self'.rev` / `self'.dirtyRev`), so **there is no version
constant to bump**. goreleaser injects the tag via `-ldflags -X`.

## Pre-flight (on master, all green)

```bash
nix flake check                       # flake apps + devshell eval
nix run .#test                        # full suite with -race
nix run .#lint                        # go vet + build + golangci-lint
nix run .#test-integration            # opt-in tests against the real NPM registry
GOEXPERIMENT=jsonv2 govulncheck ./... # dependency vulnerabilities
GOEXPERIMENT=jsonv2 go mod tidy -diff # dependency hygiene
```

All of these run in CI on push; a tag should never be cut while CI is red.

## Cut the release

1. **Finalize the changelog.** Move everything from `CHANGELOG.md`
   `[Unreleased]` into a new `## [X.Y.Z] - YYYY-MM-DD` section and commit:

   ```bash
   git switch master
   git pull
   # edit CHANGELOG.md
   git commit -am "Release vX.Y.Z"
   ```

2. **Tag (annotated, never lightweight):**

   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin master vX.Y.Z
   ```

3. **Automation takes over.** Two workflows run on the tag push:
   - **Release** (`.github/workflows/release.yml`): goreleaser builds
     static binaries for Linux and macOS (amd64 + arm64) with
     `GOEXPERIMENT=jsonv2`, publishes them to a GitHub Release with
     `SHA256SUMS`, and appends (never replaces) the release notes.
   - **CI** (`.github/workflows/ci.yml`): not triggered by tags, but already
     ran on the release commit via master.

4. **Verify** the GitHub Release exists with 4 archives + checksums, and
   that a locally built binary reports the new version:

   ```bash
   gh release list
   gh release view vX.Y.Z
   nix run . -- --version   # nix reports the short git revision, not the tag
   goreleaser release --snapshot --clean   # optional dry-run BEFORE pushing the tag
   ```

## Legacy `2.x` git tags (2020–2023)

The repository's JavaScript era (2015–2023) produced plain `2.x.y` tags
(`2.9.2`, ...). They are **not** part of the Go release line, but they sort
above `v1.x` in plain `git tag` output and may appear as "latest" in tools
that sort lexically. Keep in mind:

- Use `git tag --sort=-creatordate` to see the actual release order.
- GitHub Releases are the source of truth for "latest" (the Releases page
  ignores the JS-era tags because they have no Release objects).
- Do **not** delete the old tags — deleting is irreversible, and nothing
  breaks while they exist.

## Rollback / botched release

Tags are immutable once public. If a release is broken:

1. Delete the GitHub Release (keeps the tag), fix on master, cut `vX.Y.Z+1`.
2. Retagging the same version is a last resort — it poisons the Go module
   proxy and requires retracting the version; avoid it.

If the tag itself was never fetched by anyone (seconds old), deleting the
tag and re-tagging is acceptable.
