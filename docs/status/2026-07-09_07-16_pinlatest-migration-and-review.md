# Status Report — 2026-07-09 07:16

> **Current Status (reviewed 2026-07-09):** Historical snapshot of the
> `--pinLatest` migration session. The feature is DONE and shipping.
>
> **Completed since this report (from the 50-item list):**
>
> - CHANGELOG.md `[Unreleased]` — DONE (#1)
> - `config_test.go` — DONE: `-P`, `--pin-latest`, `PinLatest` default assertion (#2, #3)
> - AGENTS.md — DONE: execution pipeline, domain concepts, gotchas updated (#4, #5)
> - Copyright year fix — DONE: `2015-2026` (#7)
> - `latestReplaceRe` consolidation — DONE: removed; `resolveSpecVersion` sets `SNew = VNew` directly (#8)
> - `makezero` warnings — RESOLVED (#17, #18)
> - `doc.go` example — needs updating again (signatures changed in error-handling overhaul)
>
> **Rejected / Obsolete:**
>
> - `BuildManifest` options struct refactor — **REJECTED**: positional `bool` remains; YAGNI for a CLI with no external library consumers (#20)
> - Golden file tests vs `rse/upd` — **REJECTED**: no value for a Go rewrite with different output format (#25)
> - Shell completions — not done (#47, #48)
>
> ~~**Still relevant:**~~ all four resolved since (docs-health pass 2026-09-25):
>
> - ~~golangci-lint in CI — still only `go vet` (#21)~~ done at `e64d3a7` (D25/D26)
> - ~~Retry logic for 429/5xx — not done (#40)~~ done at `e64d3a7` (D28)
> - ~~`--dry-run` alias for `--nop` — not done (#43)~~ done at `e64d3a7` (D32)
> - ~~`NO_COLOR` env var support — not done (#50)~~ done at `e64d3a7` (D31)

## Session Scope

Two tasks were requested and executed:

1. **Review the whole project** and clarify in README.md that this is a Go rewrite of `https://github.com/rse/upd`
2. **Migrate PR #10** (`--pinLatest` flag) from `rse/upd` into this Go port

---

## A. FULLY DONE

### Task 1: README Origin Attribution

- Added a prominent blockquote at the top of `README.md` linking directly to `https://github.com/rse/upd`
- Rewrote the **Origin** section with explicit links to the GitHub repo and pnpm package
- Uses the word "rewrite" (not vague "port") and names both authors with links

### Task 2: `--pinLatest` Feature Migration

All code changes are complete, build + vet + tests (40/40) + race detector pass:

| File               | Change                                                                                                            |
| ------------------ | ----------------------------------------------------------------------------------------------------------------- |
| `config.go`        | `PinLatest bool` field on `Config`, `-P`/`--pin-latest` flag registration, usage strings updated                  |
| `manifest.go`      | `IsLatest bool` field on `Spec`, `latestRe` regex, `pinLatest` parameter on `BuildManifest`, classification logic |
| `engine.go`        | `latestReplaceRe` for case-insensitive `"latest"` replacement, `shouldUpdate` short-circuits for `IsLatest` specs |
| `cmd/upd/main.go`  | Passes `cfg.PinLatest` to `BuildManifest`                                                                         |
| `doc.go`           | Updated example call signature                                                                                    |
| `README.md`        | Added `-P` to usage line and flag documentation                                                                   |
| `manifest_test.go` | `TestBuildManifestPinLatest` — tests both enabled/disabled, case-insensitive matching                             |
| `engine_test.go`   | `TestEnginePinLatest` — full integration test (fetch + apply + byte-level write verification)                     |
| `engine_test.go`   | `TestEnginePinLatestDisabled` — confirms `"latest"` stays skipped without the flag                                |
| `render_test.go`   | All `BuildManifest` calls updated for new signature                                                               |
| `config_test.go`   | _(NOT updated — see section B)_                                                                                   |

**Verification:**

```
go build ./cmd/upd   → OK
go vet ./...         → OK
go test ./... -race  → OK (40 tests, 1.019s, race-clean)
```

---

## B. PARTIALLY DONE / FORGOTTEN

These are things I **should have done** as part of the `--pinLatest` migration but **did not**:

1. ~~**CHANGELOG.md `[Unreleased]` is empty** — The `--pinLatest` feature is not documented there. This is a clear omission.~~ done — documented in `[1.1.0]`
2. ~~**config_test.go not updated** — The new `-P`/`--pin-latest` flag has **zero test coverage** at the flag-parsing level: ...~~ done at `e64d3a7` era (see the 09-08 report's self-identified fixes)
3. ~~**AGENTS.md not updated** — The execution pipeline section (step 5: "Classify") and domain concepts section don't mention the `pinLatest` flag, the `IsLatest` field on `Spec`, or the `latestRe` regex. A future AI session would have an incomplete mental model.~~ done (AGENTS.md documents all three)
4. ~~**Did not run `nix run .#lint`** — AGENTS.md specifies this as the canonical lint command. I only ran `go vet` and `go build`. The project has `.golangci.yml` with 100+ linters that may surface issues (e.g., `exhaustruct` on the new `PinLatest` field).~~ done — lint is now part of the gate and runs 0 issues

---

## C. NOT STARTED

Nothing in the original two-task scope was left unstarted.

---

## D. TOTALLY FUCKED UP

Nothing was broken. No regressions. All 40 tests pass including race detector.

---

## E. WHAT WE SHOULD IMPROVE

### Architectural / Design Issues

1. ~~**Split-brain "latest" regexes** — Two separate regex patterns exist for the same concept: ...~~ done — `latestReplaceRe` removed; `resolveSpecVersion` sets `SNew = VNew` directly (09-08 report, self-identified fix #2)
2. ~~**`BuildManifest` signature change is a breaking API change** — Adding `pinLatest bool` as a third parameter breaks any external caller. An options struct or variadic options pattern would be more future-proof. (Low impact: this is a single-binary CLI with no known external library consumers.)~~ Won't implement — TODO_LIST R2
3. ~~**`shouldUpdate` special-casing** — The `IsLatest` check bypasses semver comparison entirely, always returning `true`. ... but it's worth documenting why.~~ done — documented in AGENTS.md ("kept vs updated" + Upgradable vs Skipped sections)
4. ~~**3 pre-existing `makezero` lint warnings** in `diff.go` and `render.go` (not introduced this session, but still present).~~ done — resolved; lint is 0 issues
5. ~~**Copyright year is stale** — `PrintVersion` in `config.go` says `2015-2025` but the original `rse/upd` now says `2015-2026` and today's date is 2026-07-09.~~ done at `55dfe76` era
6. ~~**No `-P` in `TestParseFlagsShortFlags`** — The flag-parsing tests are the regression net for CLI behavior; the new flag is invisible to them.~~ done at `e64d3a7` era
7. **`doc.go` example hardcodes `false`** — Doesn't show a library consumer how to enable `pinLatest`. ← open (minor)
8. ~~**README "Development" section** lists plain `go build`/`go test`/`go vet` instead of `nix run .#build` etc. (Pre-existing, but AGENTS.md says Nix-first.)~~ done — README Development section shows the nix commands

---

## F. NEXT 50 THINGS TO DO

### Immediate (this session's loose ends)

1. ~~Update `CHANGELOG.md` `[Unreleased]` with `--pinLatest` feature~~ done — in `[1.1.0]`
2. ~~Add `-P` to `TestParseFlagsShortFlags` and `--pin-latest` to `TestParseFlagsLongFlags`~~ done at `e64d3a7` era
3. ~~Add `PinLatest` default-false assertion to `TestParseFlagsDefaults`~~ done at `e64d3a7` era
4. ~~Update `AGENTS.md` execution pipeline with `pinLatest` classification step~~ done
5. ~~Update `AGENTS.md` domain concepts with `IsLatest` field and `latestRe`~~ done
6. ~~Run `nix run .#lint` and fix any new lint findings~~ done — 0 issues
7. ~~Fix copyright year `2015-2025` → `2015-2026` in `PrintVersion`~~ done at `55dfe76` era
8. ~~Unify or anchor the `latestReplaceRe` in `engine.go`~~ done — removed entirely

### Testing Improvements

9. Add test for `IsLatest` + `Nop` interaction (should count as update but not write)
10. Add test for `IsLatest` + `Greatest` mode (should resolve to greatest, not latest dist-tag)
11. Add test for `"latest"` with whitespace padding (e.g., `" latest "`)
12. Add test for `IsLatest` when registry returns error (should become `StateError`)
13. Add test for mixed manifest: some `"latest"`, some semver, some skipped — all in one pass
14. Add test for `"latest"` appearing in multiple sections (dependencies + devDependencies)
15. Add test verifying byte-level formatting preservation when replacing `"latest"`
16. Add integration test: full CLI run with `-P` flag end-to-end

### Code Quality

17. ~~Fix 3 pre-existing `makezero` lint warnings in `diff.go` (`make([][]int, oldLen+1)` → `make([][]int, 0, oldLen+1)` pattern)~~ done
18. ~~Fix `makezero` warning in `render.go:167` (`segments` slice)~~ done
19. ~~Consider consolidating the quiet/non-quiet code branches in `main.go` (noted in AGENTS.md as duplication)~~ done at `e64d3a7` (D24)
20. ~~Consider `BuildManifest` options pattern instead of positional bools~~ Won't implement — TODO_LIST R2
21. ~~Add `golangci-lint` to CI (currently CI only runs `go vet`)~~ done at `e64d3a7`
22. Consider using `structs` with exhaustive initialization to satisfy `exhaustruct` linter ← open

### Feature Parity with rse/upd

23. ~~Evaluate issue #11 (VHS demos) — decide yes/no~~ done — adopted (v1.1.0)
24. Check if there are other open PRs or issues on `rse/upd` worth migrating ← open
25. ~~Verify CLI output matches `rse/upd` byte-for-byte for identical inputs (golden file tests)~~ Won't implement — TODO_LIST R5
26. Compare the JS version's error messages with Go port's error messages for consistency ← superseded — Go messages now come from `messages.go` templates (`3cd313e`)
27. ~~Check if `rse/upd` has any config-file support beyond the `upd` field in `package.json`~~ moved to ROADMAP.md theme 2

### Documentation

28. Add `--pin-latest` to the usage examples in README
29. Add a "Why?" section to README explaining the motivation (reproducible builds)
30. Update README "Development" section to show `nix run` commands
31. Add `CONTRIBUTING.md` note about `nix run .#lint` before submitting PRs
32. Add architecture diagram (D2) to docs/
33. Update `docs/DOMAIN_LANGUAGE.md` with `pinLatest` and `IsLatest` concepts

### DevOps / CI

34. ~~Add `golangci-lint` step to `.github/workflows/ci.yml`~~ done at `e64d3a7`
35. ~~Add `go test -race` to CI (currently only `go test`)~~ done — CI test step uses `-race`
36. ~~Add release workflow with `ldflags` version injection~~ moved to TODO_LIST.md #3
37. ~~Add `nix flake check` to CI~~ moved to TODO_LIST.md #2
38. ~~Consider Renovate/Dependabot for Go dependency updates~~ done — `.github/dependabot.yml` (verified 2026-09-25)
39. Add `.golangci.yml` sarif output for GitHub Security tab ← open

### Robustness

40. ~~Add HTTP retry logic for transient NPM registry failures~~ done at `e64d3a7` (D28)
41. ~~Add `--registry` flag for custom/private NPM registry URL~~ done at `e64d3a7` (D29)
42. ~~Add timeout flag for individual package fetches~~ done at `e64d3a7` (D33)
43. ~~Add `--dry-run` alias for `--nop` (more conventional name)~~ done at `e64d3a7` (D32)
44. ~~Handle scoped packages with special characters in registry URL encoding~~ done at `e64d3a7` (`TestScopedPackageURLEncoding`)
45. ~~Add `--format json` output mode for CI/scripting~~ done at `e64d3a7` (D34, as `--json`); a `--format` flag remains an idea (TODO_LIST.md #11)
46. ~~Add exit code differentiation (0 = no updates, 1 = error, 2 = updates applied)~~ done differently — 0/1/75 semantic via `errorfamily` (documented in README)

### Polish

47. Add shell completion generation (`--bash-completion`, `--zsh-completion`)
48. Add `man` page generation
49. Add progress bar ETA calculation
50. Add color detection (respect `NO_COLOR` env var, not just `-C` flag)

---

## G. TOP 2 BLOCKING QUESTIONS

### Q1: Should `BuildManifest` use an options struct instead of a positional bool?

Adding `pinLatest bool` as a third parameter to `BuildManifest` works but is fragile — every future flag that affects classification (e.g., `--pinGreatest`, `--allowDowngrade`) would require another positional parameter or another breaking change. An `Options` struct or functional options pattern would future-proof the API. **However**, there are no known external library consumers, so the blast radius is zero today. Should I refactor now or defer?

### Q2: Should the `latestReplaceRe` regex be anchored?

The replacement regex `(?i)latest` in `engine.go` is unanchored — it replaces any occurrence of "latest" in the constraint string. Since detection only matches bare `"latest"` (via the anchored `latestRe`), the unanchored replacement is technically safe today. But it's a latent bug if someone manually sets `SOld` to something containing "latest" in a version prerelease tag (e.g., `"1.0.0-latest"`). Should I anchor it to `(?i)^latest$` for safety, or is this YAGNI?
