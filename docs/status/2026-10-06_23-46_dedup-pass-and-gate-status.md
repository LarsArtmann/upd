# Status Report — Dedup Pass (art-dupl) + Gate Repair

**Repo:** `upd` (github.com/LarsArtmann/upd) · **Branch:** `master` · **Generated:** 2026-10-06 23:46 CEST
**Scope:** This session only — deduplication of the 4 actionable art-dupl clone groups, plus what the quality gates surfaced along the way. No other research performed (per operator instruction).

---

## 0. Session Timeline (evidence trail)

| # | Event | Result |
|---|-------|--------|
| 1 | Loaded `deduplicate-code` skill, read all 4 clone sites + classification code + tests | Context complete |
| 2 | Judgment pass: 2 clones eliminated, 2 accepted | See §a/§e |
| 3 | Edited `manifest.go`, `npm.go`, `config.go` (stdlib `slices`/`maps`), `engine.go` (`shouldUpdate` via LSP symbol replace) | 4 files |
| 4 | `GOEXPERIMENT=jsonv2 go build/vet/test` | All green |
| 5 | `buildflow --build-mode dev` | **FAILED** findings gate: erraudit 1 error finding |
| 6 | Triaged finding → `aggregateFailures` (`cmd/upd/main.go:159`) `context_loss` false positive → suppressed inline with rationale | erraudit step green |
| 7 | Re-ran `art-dupl --sort total-tokens -t 1 --suggest-generics` | **4 → 2** actionable groups (198 → 196 total) |
| 8 | Full `buildflow --build-mode dev` re-run | Exit 0, gate clean |
| 9 | `nix run .#test` (race included) | PASS (all packages) |
| 10 | `nix run .#lint` | build OK, **0 issues** |
| 11 | Noticed 3 markdown files changed by no hand of mine → inspected diffs → pure table-padding realignment from BuildFlow's format step → left untouched (safety rule: not mine to revert) | Documented |
| 12 | `jscpd --min-lines 5 --min-tokens 40` | **14 clones — unchanged** (test-file duplication, tracked as TODO_LIST #11) |
| 13 | AGENTS.md updated: ERRAUDIT bullet (14th suppression) + new ART-DUPL bullet (kept-clone rationale) | Memory current |

**Files touched this session:** `config.go`, `manifest.go`, `npm.go`, `engine.go`, `cmd/upd/main.go`, `AGENTS.md` — plus 3 formatter-touched markdown files (`FEATURES.md`, `TODO_LIST.md`, `docs/status/archived/README.md`, **not authored by me**).
**Net Go diff:** ~+30/−60 lines (duplication removed, no behavior change).

---

## a) FULLY DONE ✅

| # | Work | Evidence |
|---|------|----------|
| A1 | **Clone 1 eliminated — map-keys→slice, 3 sites.** `Manifest.SortedNames` (manifest.go:128), `Packument.VersionKeys` (npm.go:298), `deprecatedFlagNames` (config.go:392) now use stdlib `slices.Sorted(maps.Keys(...))` / `slices.Collect(maps.Keys(...))`. art-dupl found 2 of the 3 sites; the npm.go twin was found by reading and killed too. | art-dupl group gone; behavior identical |
| A2 | **Clone 4 eliminated — `shouldUpdate` `StateKept` ×2 merged** into a single decision point (engine.go:207): update iff version changed AND (`IsLatest` OR semver-greater). Truth table provably unchanged — `latest` specs carry `VOld="latest"`, so the equality guard and the `IsLatest` short-circuit can never collide. | Full suite + race PASS |
| A3 | **erraudit false positive triaged and suppressed** (`cmd/upd/main.go:159`): `errCount` IS attached via `ErrPartialFailure.WithContextf("error_count", ...)`; the analyzer cannot track context through the `errs` slice into `errors.Join`. 14th documented `//nolint:erraudit` suppression, rationale in code + AGENTS.md. | buildflow gate exit 0 |
| A4 | **AGENTS.md memory updated in-session** (per aggressive-update protocol): ERRAUDIT bullet extended, new ART-DUPL bullet records both kept micro-clones with rationale so no future session re-litigates them. | 2 bullets |
| A5 | **Full verification matrix:** build+vet+test, buildflow dev gate (twice: fail→fix→pass), `nix run .#test` (race), `nix run .#lint` (0 issues), art-dupl re-run, jscpd check | All green at session end |
| A6 | **Skill compliance:** dedup-code loaded before any edit; buildflow loaded before running the quality gate (delegate-don't-duplicate honored — no raw golangci-lint invocation). | — |
| A7 | **Un-owned changes respected:** markdown realignment diffed, judged formatter output, NOT reverted. | §11 above |

## b) PARTIALLY DONE 🟡

| # | Work | Done | Missing |
|---|------|------|---------|
| B1 | **"Iterate to zero harmful clones" (dedup-skill contract)** | art-dupl `-t 1`: harmful clones gone, 2 remaining groups are documented accepts | jscpd's 14 test-file clones (110 lines, TODO_LIST #11) not touched this session — the broader repo dedup program is still open |
| B2 | **`shouldUpdate` refactor safety** | Existing suite covers kept / updated / error / nop / pin-latest paths — all PASS | No *targeted* table-driven test pinning the merged condition's 4 branches (esp. `IsLatest` × older / `IsLatest` × newer); refactor relies on incidental coverage |
| B3 | **Gate hygiene** | Findings gate green | Buildflow reported "9 tools unavailable (health check failed)" and erraudit summary shows `filesScanned: 0` — neither investigated (see §d, §g) |

## c) NOT STARTED 📋

- **jscpd re-elimination** (TODO_LIST #11): 14 clones / 110 lines / 1.61% — untouched this session (out of the art-dupl scope I was given).
- **HARVEST of this report's §f into `TODO_LIST.md`** (docs-health HARVEST) — deliberately deferred: operator said "then wait for instructions".
- **Pin-latest test hardening** (TODO_LIST #5): now *more* relevant — my `shouldUpdate` merge makes those 6 interaction cases the natural regression net.
- **Deterministic flag-suggestion order**: `deprecatedFlagNames()` iterates a map (nondeterministic); tie-breaking in `suggestFlag` is therefore unstable. Preserved as-is (behavior-preserving choice), not fixed.
- **dprint automated gate**: still absent (FEATURES.md row remains 🟡 `PARTIALLY_FUNCTIONAL`).
- **Local coverage-gate / `nix flake check` runs**: not executed this session (CI owns both; no nix files touched).
- Older TODO_LIST items glimpsed in the formatter diff (#2 goreleaser signing, #4 scoped registries, #6 malformed-`upd`-field test, #7 retry-path micro-gaps, #8 401/403 auth classification): untouched.

## d) TOTALLY FUCKED UP 💥

Honest list — nothing catastrophic, but these are real:

| # | Item | Severity | Detail |
|---|------|----------|--------|
| D1 | **The gate-clean contradiction is unresolved.** AGENTS.md claims erraudit "gate-clean as of 2026-10-06", yet today's live gate flagged `aggregateFailures`. Either this morning's claim was unverified/cached, or the tool changed mid-day. I suppressed the finding without root-causing the *discrepancy itself* — that violates "fix problems at root cause". A memory file whose "clean" claims can silently disagree with the live gate is a trust problem for every future session. | 🔴 High | §g Q1 |
| D2 | **First `main.go` edit had broken indentation** (comment lines written without the leading tab; caught on read-back and fixed). Small, but it was a quality slip in the one file I edited under time pressure. | 🟡 Low | Fixed same-session |
| D3 | **Buildflow health:** "9 tools unavailable (health check failed)" in both dev runs — unexamined. If erraudit itself was one of the 9, the "clean" verdicts are weaker than they look. | 🟠 Medium | Next-step N5 |
| D4 | **Working tree is uncommitted** (6 authored files + 3 formatter files). The pma auto-commit daemon has a documented blind spot; until a commit lands, the session's work exists only in the working tree. Operator has not said "commit" — so this is a flag, not an action. | 🟡 Low | Next-step N2 |

## e) WHAT WE SHOULD IMPROVE 🔧

1. **Verify-then-claim for AGENTS.md gate bullets** — record *which* binary/step/cache state produced a "clean" verdict, so claims are reproducible (would have caught D1 immediately).
2. **Refactor + test in the same breath** — when merging condition branches, add the truth-table test before claiming "truth table unchanged", instead of leaning on the existing suite (B2).
3. **Root-cause gates, don't just silence them** — the nolint was correct, but the *discrepancy* (D1) deserved a `git log`/cache-bypass dig before moving on.
4. **Inline rationale for accepted clones** — the dedup skill says "leave a one-line rationale" in code; I documented in AGENTS.md instead (repo no-comment convention). Deviation is defensible but should be a *policy*, not an accident (§g Q2).
5. **Fix edit hygiene** — exact-indentation discipline on comment insertions (D2 was avoidable with a fuller old_string).
6. **Health-check the gate runner first** — a "9 tools unavailable" banner should be triaged (buildflow doctor) *before* trusting any green summary.
7. **Stop recurring markdown churn** — BuildFlow re-pads markdown tables every run, guaranteeing noisy diffs in every session that runs format. One-time commit of the realignment (or disable md table formatting) ends the noise permanently.
8. **Determinism by default** — map-order iteration feeding tie-breakable decisions (flag suggestions) should be sorted once, with a test, instead of carried as a known nondeterminism.

## f) NEXT — 40 things to get done (sorted by impact; session-grounded, no external research)

**Immediate / high impact**

| N# | Task | Source |
|----|------|--------|
| N1 | HARVEST this §f list into `TODO_LIST.md` / `ROADMAP.md` (docs-health HARVEST mode) | skill contract |
| N2 | Commit the session: 5 Go files + AGENTS.md (+ decide fate of the 3 formatter-touched markdown files) | §d D4 |
| N3 | Add table-driven test pinning `shouldUpdate`'s 4 branches: equal→kept, `IsLatest`+older→update, `IsLatest`+newer→update, plain older→kept | B2 |
| N4 | Root-cause the erraudit gate-clean discrepancy: re-run with `BUILDFLOW_NO_RESULT_CACHE=1`, diff erraudit binary version vs this morning's run, then correct or confirm the AGENTS.md claim | D1 |
| N5 | `buildflow doctor` — resolve "9 tools unavailable (health check failed)" | D3 |
| N6 | Investigate erraudit finding metadata `filesScanned: 0` — did it actually scan all 33 files, or only cache-miss files? | B3 |
| N7 | One-time commit of markdown table realignment (or configure it away) to stop per-session format churn | e7 |
| N8 | Pin-latest test hardening (TODO_LIST #5: `IsLatest` × `--nop`, × `--greatest`, padded `" latest "`, registry-error, multi-section, CLI `-P` e2e) — elevated priority post-refactor | TODO_LIST #5 |

**Session follow-ups (small, this repo)**

| N# | Task | Source |
|----|------|--------|
| N9 | Decide + record canonical art-dupl threshold (`-t 1` vs `-t 5`) in AGENTS.md ART-DUPL bullet; add exact reproduce command | §g Q3 |
| N10 | Sort flag-suggestion candidates (`deprecatedFlagNames`/`flagNames`) for deterministic tie-breaks + unit test | e8 |
| N11 | Run local coverage check to pre-verify the CI 80% gate after the refactor | c |
| N12 | Run `nix flake check` once (CI owns it; cheap local assurance) | c |
| N13 | Verify my trailing `//nolint:erraudit` placement survives a `buildflow format` pass (golines relocation gotcha documented in AGENTS.md) | A3 |
| N14 | Run `buildflow --build-mode full` once before the next release (markdown-lint + gitleaks + coverage parity with CI) | — |
| N15 | E2E smoke of the built binary on a fixture package.json with kept+updated+error mix — visual confirmation `kept` rendering is unchanged | — |
| N16 | `go test -shuffle=on ./...` to smoke out map-order-dependent tests | e8 |
| N17 | jscpd re-elimination sweep (TODO_LIST #11), then update the AGENTS.md jscpd bullet with the new count | TODO_LIST #11 |
| N18 | Update FEATURES.md rows touched by this session (dedup state; Formatting row if N7 lands) | — |
| N19 | Fold the dedup refactor into the next release notes (no behavior change; per docs/RELEASING.md ritual) | — |
| N20 | Verify pma auto-commit daemon actually picks up this session's files (buildflow anti-pattern: don't trust the daemon) | D4 |

**TODO_LIST backlog glimpsed this session (unmodified, owned elsewhere)**

| N# | Task | Source |
|----|------|--------|
| N21 | #2 Goreleaser signing/provenance (cosign keyless or checksum signatures) | TODO_LIST #2 |
| N22 | #4 Scoped registries: `@scope:registry=` + per-scope tokens in fetch path | TODO_LIST #4 |
| N23 | #6 `GetUpdArgs` malformed-`upd`-field fatal-path test | TODO_LIST #6 |
| N24 | #7 Retry micro-gaps: `parseRetryAfter` HTTP-date/invalid/past branches; ctx-cancel mid-backoff | TODO_LIST #7 |
| N25 | #8 401/403 → distinct Rejection-family sentinel (exit 1) + classification matrix incl. 410 | TODO_LIST #8 |
| N26 | dprint automated gate in flake/CI (closes the 🟡 Formatting row) | FEATURES.md |

**Tooling / upstream (BuildFlow fleet decisions)**

| N# | Task | Source |
|----|------|--------|
| N27 | Upstream erraudit limitation: context through `errors.Join(...)` slices reported as `context_loss` (linter-building territory) | A3 |
| N28 | Consider periodic art-dupl/jscpd dedup step in BuildFlow (fleet-wide value → upstream, not per-repo script) | buildflow skill |
| N29 | Capture `buildflow timings --regressions` baseline for upd | buildflow skill |
| N30 | Record in AGENTS.md that `art-dupl` HTML mode (`--html`) is available for visual passes | dedup skill |

**Nice-to-have polish**

| N# | Task | Source |
|----|------|--------|
| N31 | Fuzz/table-test `versionIsGreater` semver edges (prerelease, build metadata, unparseable→conservative-true) | session read |
| N32 | Micro-bench `slices.Sorted(maps.Keys())` vs old loops on large manifests (expect negligible; confirm and close) | A1 |
| N33 | Add the two accepted-clone rationales inline (1-line comments) if policy lands on inline (§g Q2) | §g Q2 |
| N34 | Re-run art-dupl with `--html` next pass for the visual artifact | N30 |
| N35 | Sweep AGENTS.md for other same-day "clean" claims and stamp each with its verification command | D1 |
| N36 | Consider `git town` flow for the next change set (repo configures git-town.toml; this session worked directly on master) | AGENTS.md |
| N37 | After N7: confirm buildflow format step is idempotent (second run → zero markdown diffs) | N7 |
| N38 | Document the `nolint_filter` warning ("unknown linters: erraudit") in AGENTS.md as expected noise for standalone golangci-lint runs | observed |
| N39 | Add a flake app (e.g. `.#dupl`) wrapping the canonical art-dupl invocation so future passes share flags | N9 |
| N40 | Post-release: re-verify all four demo GIFs still match CLI behavior (dedup refactor shouldn't change output, but the ritual is cheap) | AGENTS.md |

## g) Questions I can NOT figure out myself (3)

1. **Erraudit gate-clean discrepancy (D1):** AGENTS.md says erraudit was gate-clean "as of 2026-10-06", yet today's live gate flagged `aggregateFailures`. What exactly was run this morning when that claim was written — which buildflow/erraudit binary version, and with or without result cache? (This decides whether I should distrust other same-day "clean" claims and whether N4 is a memory bug or a tool regression.)
2. **Accept-rationale policy:** for clones deliberately kept, do you want the one-line rationale **inline in code** (dedup-skill letter, violates your no-comment default) or **AGENTS.md-only** (what I did)? I'll follow whichever you call canonical, repo-wide.
3. **Canonical art-dupl threshold:** your paste used `-t 1`; the skill defaults to `-t 5`. Which should future dedup passes in this repo treat as the baseline? (AGENTS.md's new ART-DUPL bullet currently describes the `-t 1` pass without mandating it.)

---

*Point-in-time snapshot — goes stale. When bringing this current later, use docs-health ANNOTATE (inline correction or end-of-file appendix), never rewrite. Section (f) is the HARVEST input for `TODO_LIST.md` (N1).*
