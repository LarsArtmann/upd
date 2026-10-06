# Status Report: 2026-07-09 09:08 — json/v2 Migration, VHS Demos & Cleanup

> **Current Status (reviewed 2026-07-09):** Historical snapshot. All items in
> section A (FULLY DONE) remain accurate and shipping.
>
> **Completed since this report:**
>
> - `makezero` lint warnings — RESOLVED (#5 in section B)
> - Error handling overhaul — DONE: `Spec.Err`, `ErrRegistryUnavailable`, exit
>   code 75, warnings pipeline, error detail block in renderer
> - `GetDependencySection` and `GetUpdArgs` now return errors (no more silent swallowing)
> - `BuildManifest` now returns `(Manifest, []string)` with warnings
>
> **Still open from this report:** → all resolved or routed (docs-health pass 2026-09-25):
>
> - ~~VHS demo was never actually rendered — infrastructure exists but unverified (#1 in section B)~~ done — rendered and published to `vhs.charm.sh` (v1.1.0, `2332acb` era); README embeds the GIF. Re-render for the fang-styled CLI tracked as TODO_LIST.md #10
> - ~~golangci-lint not in CI — still only `go vet` (#3 in section C)~~ done at `e64d3a7` (D25)
> - ~~`nix run .#lint` still only runs `go vet`, not `golangci-lint` (#3 in section E)~~ done at `e64d3a7` (D26)
> - ~~`CONTRIBUTING.md` has been updated with `GOEXPERIMENT=jsonv2` (#8 in section E)~~ done (`32c208b`)
> - ~~`BuildManifest` options struct — REJECTED (YAGNI) (#4 in section C)~~ confirmed rejected (TODO_LIST R2)
>
> ~~**Note:** `doc.go` example code has broken signatures again — the error-handling
> overhaul changed `BuildManifest` and `GetUpdArgs` return types but `doc.go` was
> not updated. This is tracked as a known issue.~~ resolved at `9ca148e`

---

## A. FULLY DONE (verified green)

### 1. json/v2 + jsontext Migration

- **pnpm.go**: Replaced `gjson.GetBytes` path queries with typed struct unmarshaling via `encoding/json/v2`. `LatestVersion()` now unmarshals into `struct { DistTags struct { Latest string } }`, `VersionKeys()` into `struct { Versions map[string]struct{} }`.
- **packagejson.go**: Replaced all `gjson` usage with `encoding/json/jsontext` streaming decoder. `UpdateDependency()` now uses `jsontext.NewDecoder` with `ReadToken()` + `ReadValue()` + `InputOffset()` for byte-precise surgical edits. `GetDependencySection()` and `GetUpdArgs()` use `json.Unmarshal` into `map[string]jsontext.Value` and typed structs respectively. JSON validation uses `jsontext.Value.IsValid()`.
- **engine_test.go**: Migrated mock registry from `encoding/json` (v1) `json.NewEncoder` to `encoding/json/v2` `json.MarshalWrite`. Fixed variable shadowing (`json` → `originalJSON`) in `TestEngineApplyUpdatesNop`.
- **go.mod**: Removed `github.com/tidwall/gjson` and its transitive deps (`tidwall/match`, `tidwall/pretty`). Down from 4 direct dependencies to 3.
- **.golangci.yml**: Removed stale `github.com/tidwall/gjson` from depguard allowlist.
- **Verification**: `GOEXPERIMENT=jsonv2 go build`, `go vet`, `go test -race` (44 tests), `nix build`, `nix run .#lint`, `nix run .#test`, `nix flake check` — all green. Zero `gjson` references, zero json v1 imports remain in the codebase.

### 2. GOEXPERIMENT=jsonv2 Wiring

- **flake.nix**: `env.GOEXPERIMENT = "jsonv2"` on `buildGoModule`. All nix apps (`build`, `test`, `lint`, `run`, `demo`) export `GOEXPERIMENT=jsonv2`. DevShell sets `GOEXPERIMENT` env var.
- **.github/workflows/ci.yml**: Top-level `env: GOEXPERIMENT: jsonv2` on the build job. Test step now uses `-race` flag.
- **flake.nix vendorHash**: Updated from stale hash to `sha256-HHBnbQrRKhy4EGNZfFyo8C7qHzhAASITutgYa4eHADU=` matching the new dependency set.

### 3. VHS Animated Demo Infrastructure

- **demo/demo.tape**: VHS tape script — shows `cat package.json`, `upd -n` (dry-run table), `upd -nP` (pin latest). Uses Catppuccin Mocha theme, 16pt font, 1000x580 dimensions.
- **demo/package.json**: Fixture with 5 outdated deps including a bare `"semver": "latest"` for pinLatest demonstration.
- **demo/README.md**: Explains local render vs cloud publish workflow.
- **flake.nix `demo` app**: `nix run .#demo` builds upd fresh, puts it on PATH, runs VHS on all `.tape` files. `nix run .#demo -- --publish` renders + publishes to `vhs.charm.sh` cloud (returns shareable URLs).
- **.gitignore**: `demo/*.gif`, `demo/*.webm`, `demo/*.mp4` — GIFs never committed.
- **README.md**: New Demo section with VHS badge, render/publish commands, pointer to tape sources.
- **devShell**: Added `vhs`, `ttyd`, `ffmpeg` to `buildInputs`.

### 4. Self-Identified Fixes from Prior Session

- **Copyright year**: `2015-2025` → `2015-2026` in `config.go:164` and `README.md:71`.
- **Latest regex consolidation**: Removed `latestReplaceRe` (unanchored, `(?i)latest`) from `engine.go`. `resolveSpecVersion` now sets `spec.SNew = vNew` directly for `IsLatest` specs — no regex replacement needed. Eliminated the two-regex inconsistency.
- **config_test.go**: Added `-P` to `TestParseFlagsShortFlags`, `--pin-latest` to `TestParseFlagsLongFlags`, `PinLatest` assertion to `TestParseFlagsDefaults`.
- **CHANGELOG.md**: `[Unreleased]` section with `--pinLatest`, json/v2 migration, VHS demos, copyright update.
- **AGENTS.md**: Fully rewritten — json/v2 + jsontext docs, GOEXPERIMENT requirement, pinLatest/IsLatest domain docs, jsontext.Token voiding gotcha, updated dependency list (3 direct), VHS demo workflow, race detector in CI.

---

## B. PARTIALLY DONE

### 1. ~~VHS Demo — Never Actually Rendered~~ done

The tape file, fixture, nix app, and documentation are all in place, but `nix run .#demo` was **never executed** during this session. VHS requires `ttyd` + `ffmpeg` + a running terminal environment that may not work headless. The tape syntax was not validated. The GIF was never produced. The cloud publish path was never tested. This is infrastructure-ready but unverified. → done — rendered + published (v1.1.0); current re-render need tracked as TODO_LIST.md #10

### 2. ~~Pre-existing Lint Warnings (3 `makezero`)~~ done

`diff.go:40`, `diff.go:42`, `render.go:167` have `makezero` warnings (slices declared with `make` but zero initial length where the linter wants pre-allocated capacity). These are **pre-existing** (not introduced this session) but remain unfixed. They would need to be addressed to get a fully clean `golangci-lint` run. → resolved (noted in the round-2-era headers; lint is 0 issues since)

---

## C. NOT STARTED

### 1. ~~VHS GitHub Action for Auto-Rendering~~ Won't implement

~~No `.github/workflows/vhs.yml` was created.~~ The `charmbracelet/vhs-action@v2` GitHub Action could automatically render tapes on push and auto-commit GIFs or publish to cloud. This would make demos self-updating. → Won't implement — GIFs stay git-ignored and are published on demand via `nix run .#demo -- --publish`; auto-committing binaries contradicts that policy.

### 2. ~~Additional Demo Tapes~~ done (2026-10-06)

~~Only one tape (`demo.tape`) exists.~~ Could add: `pin-latest.tape` (focused pinLatest demo), `greatest.tape` (`-g` flag demo), `patterns.tape` (glob filtering demo). → done — all three tapes exist in `demo/`, rendered and published; README links every GIF.

### 3. ~~`golangci-lint` Integration in CI~~ done at `e64d3a7`

CI runs only `go vet`. The project has a `.golangci.yml` with 100+ linters but `golangci-lint` is not run in CI or in `nix run .#lint`. → done — separate CI job + lint app (D25/D26); pinned v2.12.2 in `e13492e`

### 4. ~~`doc.go` Example Verification~~ done at `8f8d6fd`

The doc.go example was updated to `BuildManifest(pkg, pkg.GetUpdArgs(), false)` but the example is not compiled or tested as part of the test suite (Go doc examples with `// Output:` comments would add compile-time verification). → done — `example_test.go` compile-verifies the documented flow (`8f8d6fd`); listed in CHANGELOG as "compile-verified package example"

### 5. ~~`BuildManifest` Options Struct Refactor~~ Won't implement

`BuildManifest` takes `pinLatest bool` as a positional third arg. Every caller was updated, but the API would be cleaner with an options struct or functional options pattern. Flagged in prior session, never addressed. → TODO_LIST R2 (YAGNI)

---

## D. TOTALLY FUCKED UP (nothing)

No regressions, no broken builds, no data loss. The one issue during this session — `jsontext.Token` voiding panic in `UpdateDependency` — was diagnosed and fixed immediately (capture `keyTok.String()` before `ReadValue()`). The duplicate `err :=` compile error was also caught and fixed before pushing forward.

**One honest caveat**: The VHS demo infrastructure is "ready to run" but was never actually run. If `ttyd` fails in the nix shell, or if the tape syntax has a typo, the demo won't work. This is unverified infrastructure.

---

## E. WHAT WE SHOULD IMPROVE

1. ~~**VHS tape must be tested** — The tape file was written blind. It needs at least one render to validate syntax, timing, and visual quality.~~ done — rendered + published (v1.1.0); tapes reworked and re-rendered 2026-10-06
2. ~~**No `golangci-lint` in CI** — The project has an extensive `.golangci.yml` but it's never actually run automatically. The 3 `makezero` warnings prove it's not being enforced.~~ done at `e64d3a7`
3. ~~**`nix run .#lint` only runs `go vet`** — It doesn't run `golangci-lint` despite the config file existing. Should be added.~~ done at `e64d3a7`
4. ~~**`BuildManifest` API smell** — Three positional args (two `[]string`/`string` slices + one `bool`) is a calling-convention footgun. Options struct would prevent argument transposition bugs.~~ Won't implement — TODO_LIST R2
5. ~~**Test file shadowing risk** — The `json` variable name in a test file shadowed the `encoding/json` import, requiring a rename. With `json/v2` this is less likely (the import is `json` not `encoding/json`) but the pattern of importing `json` and then naming local variables `json` remains a trap.~~ NOT-DO — awareness note only; `encoding/json/v2` is imported as `json`, which dissolves the original collision
6. ~~**No integration test with real registry** — All tests use mock registries. A single opt-in integration test (behind a build tag or `-run=Integration`) hitting `registry.npmjs.org` would catch real-world regressions.~~ done at `8f8d6fd` — `integration_registry_test.go` + `nix run .#test-integration`
7. ~~**flake.lock changed unexpectedly** — `flake.lock` shows 18 lines changed in the diff. This was not intentional and may be a side effect of `nix` operations during the session. Should be reviewed.~~ done — reviewed; lockfile refreshes are routine since (`f020505` etc.)
8. ~~**CONTRIBUTING.md not updated** — Doesn't mention `GOEXPERIMENT=jsonv2` requirement for contributors.~~ done at `32c208b`

---

## F. Up to 50 Things to Get Done Next

### High Priority (P0)

1. ~~**Render the VHS demo** — Run `nix run .#demo` at least once to verify it works~~ done (v1.1.0)
2. ~~**Publish the demo** — Run `nix run .#demo -- --publish` to get a real `vhs.charm.sh` URL~~ done (v1.1.0)
3. ~~**Embed the published GIF URL** in README.md once obtained~~ done — README Demo section
4. ~~**Review `flake.lock` diff** — Verify the 18-line change is intentional and not drift~~ done — lockfile refreshed routinely since (`f020505` etc.)
5. ~~**Fix the 3 `makezero` warnings** in `diff.go` and `render.go` for a fully clean lint~~ done — lint is 0 issues

### Medium Priority (P1)

6. ~~Add `golangci-lint` to CI workflow (`.github/workflows/ci.yml`)~~ done at `e64d3a7`
7. ~~Add `golangci-lint` to `nix run .#lint`~~ done at `e64d3a7`
8. ~~Add `.github/workflows/vhs.yml` with `charmbracelet/vhs-action@v2` for auto-rendering~~ Won't implement — demos publish on demand; no binary artifacts in the repo
9. ~~Update `CONTRIBUTING.md` with `GOEXPERIMENT=jsonv2` requirement~~ done at `32c208b`
10. ~~Refactor `BuildManifest` to use an options struct (resolve Q1 from prior session)~~ Won't implement — TODO_LIST R2
11. ~~Add focused demo tapes: `pin-latest.tape`, `greatest.tape`, `patterns.tape`~~ done 2026-10-06 — all three tapes rendered + published
12. ~~Add an integration test (build-tagged) hitting the real NPM registry~~ done at `8f8d6fd` — `nix run .#test-integration`
13. ~~Add Go doc examples with `// Output:` to `doc.go` for compile-time example verification~~ done at `8f8d6fd` — `example_test.go`
14. ~~Pin VHS version in `flake.nix` devShell (currently unpinned via nixpkgs)~~ Won't implement — vhs enters through the flake-locked nixpkgs input, which already pins it
15. ~~Add `meta.description` to all nix apps (flake check warns about this)~~ done at `e64d3a7` (D44)

### Lower Priority (P2)

16. ~~Consider `sjson` for writes instead of full byte-splice rebuild in `UpdateDependency`~~ Won't implement — TODO_LIST R6
17. ~~Add benchmark tests comparing json/v2 vs old gjson performance~~ Won't implement — gjson was removed as a dependency; `benchmark_test.go` covers the live hot paths
18. ~~Add property-based tests for `versionRe` and `latestRe` regex edge cases~~ done at `8f8d6fd` — `manifest_property_test.go`
19. ~~Document the `jsontext.Token` voiding gotcha in a code comment near the decoder usage~~ done differently — documented as a critical gotcha in AGENTS.md (the project's canonical memory) instead of a code comment
20. ~~Add a `Makefile`-equivalent help target: `nix run .#help` or similar~~ Won't implement — `nix flake show` lists every app with its `meta.description`; a help app would duplicate that
21. ~~Consider `embedding/json/v2` `Marshalers`/`Unmarshalers` for custom Packument parsing~~ Won't implement — struct-based unmarshaling in `npm.go` covers the two shapes we need
22. ~~Add `gofumpt` to devShell and CI for stricter formatting~~ Won't implement — dprint is the formatter (`e13492e`); a second formatter would fight it
23. ~~Add `govulncheck` to CI for dependency vulnerability scanning~~ done — dedicated CI job (now non-failing and verified clean)
24. ~~Consider `reuse` compliance for SPDX license headers~~ Won't implement — single-maintainer project with a standard MIT LICENSE; tooling ceremony without payoff
25. ~~Add a `CHANGELOG.md` entry for `[1.1.0]` when ready to release~~ done — `[1.1.0] - 2026-07-16` section exists
26. ~~Tag `v1.1.0` once all P0 items are resolved~~ done at `2332acb`
27. ~~Add `--dry-run` as an alias for `-n` (discoverability)~~ done at `e64d3a7` (D32)
28. ~~Add shell completions (bash/zsh/fish) generation~~ done at `81d8c44`
29. ~~Consider a `--format json` output mode for CI/automation consumption~~ done — shipped as `--json` (v1.1.0), replaced by `--format=json` (`2ea7868`)
30. ~~Add a `upd.test` binary or test harness for E2E testing~~ done differently — CLI-level end-to-end tests live in `cmd/upd/main_test.go` (`8f8d6fd`); a separate binary adds nothing

### Polish (P3)

31. ~~Improve demo tape: add typed comments, better pacing, terminal clear~~ done 2026-10-06 — `demo.tape` reworked with fixtures
32. ~~Add a second tape showing the `--greatest` flag~~ done 2026-10-06 — `greatest.tape`
33. ~~Add a tape showing negative glob patterns (`!pattern`)~~ done 2026-10-06 — `patterns.tape`
34. ~~Add VHS theme matching the terminal colors in `render.go`~~ Won't implement — tapes use a Catppuccin palette; replicating upd's exact ANSI colors is fiddly with no demand
35. ~~Consider WebM output in addition to GIF (smaller, higher quality)~~ Won't implement — GIF is the README-embedded artifact; `.gitignore` already excludes stray webm/mp4
36. ~~Add a `demo/` entry to `.gitattributes` if needed for line-ending handling~~ NOT-DO — no line-ending issues with tape files observed
37. ~~Review all error messages for user-facing quality (What/Reassure/Why/Fix/Escape pattern)~~ done — `messages.go` registers What/Why/Fix/WayOut templates for every error code (`3cd313e`)
38. ~~Add `--registry` flag for custom/private NPM registry support~~ done at `e64d3a7` (D29)
39. ~~Consider rate-limiting backoff for 429 responses from registry~~ done — exponential backoff + `Retry-After` support (`e64d3a7`, D28)
40. ~~Add HTTP/2 support verification (Go default client should use it)~~ Won't implement — Go's default transport negotiates h2 automatically; nothing actionable
41. ~~Add a man page (`man/upd.1`)~~ done at `81d8c44` — hidden `upd man` emits roff on demand
42. ~~Add `nix run .#bench` app for running benchmarks~~ Won't implement — `go test -bench=. ./...` inside the devShell (which exports `GOEXPERIMENT=jsonv2`) is sufficient
43. ~~Consider NixOS module for running upd as a periodic systemd service~~ Won't implement — a static binary on a PATH needs no module; a plain systemd timer/cron entry calls it directly
44. ~~Add `renovate.json` or similar for self-hosted dependency updates~~ done differently — `.github/dependabot.yml` (`7a1e31f`)
45. ~~Add issue/PR templates to `.github/`~~ done at `cced077`
46. ~~Add `CODE_OF_CONDUCT.md`~~ Won't implement — single-maintainer project; CONTRIBUTING covers the basics
47. ~~Add `FUNDING.yml`~~ Won't implement — no funding channels for this project
48. ~~Review and update `docs/DOMAIN_LANGUAGE.md` with new terms~~ done — glossary carries `pinLatest`, `IsLatest`, `packument`, and the state machine (verified 2026-09-25)
49. ~~Add architecture diagram (D2 or Mermaid) to README or docs/~~ moved to ROADMAP.md theme 2
50. ~~Consider extracting diff algorithm (`diff.go`) into its own package~~ Won't implement — contradicts the single-root-package convention (AGENTS.md); no external consumer

---

## G. Top 2 Questions

### Q1: Should `nix run .#lint` run `golangci-lint` (with the full 100+ linter config) or stay as `go vet`?

The project has `.golangci.yml` with extensive linter configuration but `nix run .#lint` only runs `go vet` + `go build`. Adding `golangci-lint` would surface the 3 `makezero` warnings and enforce the config, but would also likely produce many warnings on untouched files (the AGENTS.md explicitly says "expect loud diagnostics"). Should I add it and deal with the noise, or keep `go vet` as the standard?

> **Resolved:** golangci-lint was added to both the lint app and CI at `e64d3a7`; the noise was worked down to 0 issues.

### Q2: Should the VHS demo GIFs be auto-committed by CI or always cloud-hosted?

Two approaches: (a) GitHub Action auto-renders and commits GIFs to the repo (works offline, renders in Markdown natively on GitHub), or (b) always cloud-hosted via `vhs.charm.sh` (no binary files in repo, but requires network to view and links can expire). The user asked for cloud hosting, but option (a) is the more common VHS pattern. Which is preferred?

> **Resolved:** cloud-hosted — README embeds `vhs.charm.sh` URLs; GIFs remain git-ignored and re-render locally with `nix run .#demo`.
