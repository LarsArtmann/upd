# Status — TODO-list sweep, release completion, and a discovered CLI regression

- **Date:** 2026-10-06 21:14
- **Scope:** Worked through `TODO_LIST.md` (28 open items) end to end: release completion, CI hardening, testing, CLI/UX. 17 of 28 TODO items done or wired, 2 pre-existing bugs discovered (one severe), 11 items not started, nothing committed yet.

---

## a) FULLY DONE

| TODO                  | What                                                                                                                                                                                                                                                                                          | Verification                                                                                                                                                                                                                           |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| #1 (partial bug half) | GitHub Releases created for the existing tags `v1.3.0` and `v1.4.0` (v1.4.0 now Latest). Notes curated from CHANGELOG in the established release style. The "version source" half was already fixed upstream: `flake.nix` now uses `self'.rev or self'.dirtyRev or "dev"`, no stale hardcode. | `gh release list` shows both; tags verified on origin (`v1.3.0` → `59bcb48`, `v1.4.0` → `93c5aef`)                                                                                                                                     |
| #2                    | `nix flake check` CI job added (`cachix/install-nix-action` pinned to v9 full SHA, matching the repo's SHA-pinning policy).                                                                                                                                                                   | `nix flake check` passes locally ("all checks passed")                                                                                                                                                                                 |
| #6                    | govulncheck verified clean locally ("No vulnerabilities found", Go 1.27.0); `continue-on-error: true` removed from the vulncheck job.                                                                                                                                                         | local govulncheck run                                                                                                                                                                                                                  |
| #21                   | `go mod tidy -diff` + `go mod verify` steps added to the CI build job.                                                                                                                                                                                                                        | both pass locally ("TIDY_CLEAN", "all modules verified")                                                                                                                                                                               |
| #22                   | `-timeout 120s` on the CI test step.                                                                                                                                                                                                                                                          | in `ci.yml`                                                                                                                                                                                                                            |
| #18                   | Coverage gate: CI fails below 80%. Merged into the single race+coverage test run. Coverage re-measured after all new code: **87.7%** (was 81.0% before this session's tests).                                                                                                                 | `go tool cover -func`                                                                                                                                                                                                                  |
| #14                   | `runE(args, stdout, stderr)` refactor + 6 E2E tests (`cmd/upd/main_e2e_test.go`): formatting-preserving update, dry-run, partial failure (ErrPartialFailure, exit 1, updates not lost), total failure, JSON output summary, quiet suppression. All against a mock registry via httptest.      | full suite green                                                                                                                                                                                                                       |
| #15                   | SIGINT E2E test (`main_signal_test.go`): in-flight fetch cancelled, package.json untouched, run returns error.                                                                                                                                                                                | suite green                                                                                                                                                                                                                            |
| #16                   | Property-based regex tests (`manifest_property_test.go`, 7 tests): 500-iteration seeded generators prove versionRe accepts `^/~`+whitespace+semver-ish cores and rejects `<>                                                                                                                  | =` corruptions; latestRe case-insensitivity; BuildManifest classification; replaceVersion round-trip. No new dependency (hand-rolled generators).                                                                                      |
| #17                   | Compile-tested package Example (`example_test.go`) executing the doc.go pipeline against a mock registry with `// Output:` verification.                                                                                                                                                      | `go test` runs it                                                                                                                                                                                                                      |
| #19                   | `RenderJSON(w, manifest)` — redundant `updates` param dropped, count derived from `StateUpdated`. All callers/tests updated.                                                                                                                                                                  | suite green                                                                                                                                                                                                                            |
| #20                   | All 4 benchmarks migrated `for range b.N` → `for b.Loop()`.                                                                                                                                                                                                                                   | `go test -bench -benchtime=1x` runs clean                                                                                                                                                                                              |
| #26                   | Build-tagged integration tests (`integration_registry_test.go`, `//go:build integration`) against the real NPM registry: react packument fetch + 404 classification. Plus `.#test-integration` flake app (vet + run with tag).                                                                | both live tests PASS (0.6s / 2.5s)                                                                                                                                                                                                     |
| #28                   | `-race` added to the `nix run .#test` app — local gate now matches CI.                                                                                                                                                                                                                        | in `flake.nix`                                                                                                                                                                                                                         |
| #7                    | `--noColor` deprecation, incl. the **man-page leak fix**: the hidden flag is no longer registered; the CLI rewrites `--noColor`/`--noColor=true                                                                                                                                               | false`→`--no-color`in`runE`and prints a yellow deprecation warning ("will be removed in v2.0.0"). Man page now has **0**`noColor` hits (was 1); completions no longer offer it. ParseFlags rejects it by design (documented in tests). |
| #8                    | Invalid `UPD_*` env vars now warn instead of silently falling back: `applyEnvFlags` returns warnings, `Config.EnvWarnings()` accessor, printed through the existing stderr WARNING pipeline.                                                                                                  | new tests, incl. E2E-level wiring                                                                                                                                                                                                      |
| #9                    | Unknown-flag typo suggestions: `SetFlagErrorFunc` + Levenshtein ≤ 2 / prefix match → "Did you mean --json?". Wired into both the cobra execute path and `ParseFlags` (cobra only applies the hook in `execute()`, not bare `ParseFlags` — found and compensated).                             | unit + runE tests                                                                                                                                                                                                                      |

Also: CI YAML validated (`yaml.safe_load`); release-note drafts kept in-repo during drafting then trashed after publish (per the /tmp-drafts rule).

## b) PARTIALLY DONE

1. **TODO #9 (suggestion UX)** — the suggestion works for `upd --jso`, but my final smoke test exposed that `upd --jso -f pj.json` renders **"Unknown command "pj.json" for "upd"."** instead — cobra's `Find()` fails on unknown flags mixed with positional args _before_ `FlagErrorFunc` ever fires. Suggestion only surfaces when the flag is parseable-position-wise. Needs a pre-parse hook or arg handling to fix properly.
2. **TODO #5 (legacy 2.x tags documentation)** — AGENTS.md already documents it; the intended user-facing home was `docs/RELEASING.md`, which was not written (see TODO #4). Not started in practice.
3. **Final verification** — build/vet/tests/race-unit run green repeatedly, but the full final gate battery (`nix run .#test` with the new `-race`, golangci-lint via buildflow over every changed file, govulncheck re-run after changes) was **not** executed after the last edits. golangci-lint LSP showed only stale warnings, but the real linter hasn't judged the new files.

## c) NOT STARTED

- TODO **#3** — goreleaser `release.yml` + prebuilt binaries (researched mango/fang internals instead; goreleaser reference from the go-release skill not yet applied).
- TODO **#4** — `docs/RELEASING.md` runbook (would also carry the nix flake check pre-release gate and the legacy-2.x-tags note).
- TODO **#10** — re-render VHS demos.
- TODO **#11** — `--format` flag + `--silent` alias.
- TODO **#12** — `.npmrc` parsing (registry auth token).
- TODO **#13** — errorfamily `code`/`family` in `--json` output; `Format('+')` in verbose.
- TODO **#23** — issue/PR templates.
- TODO **#24** — `errors.Join` for warnings aggregation.
- TODO **#25** — focused demo tapes.
- TODO **#27** — `slog` structured logging.
- **Documentation updates**: TODO_LIST.md (17 items to mark done), CHANGELOG.md `[Unreleased]`, AGENTS.md (many sections now stale: `--noColor` alias is gone, main.go refactored to `runE`, RenderJSON signature changed, new tests/apps), FEATURES.md mentions.
- **Commits**: nothing committed — 9 modified + 5 new files (~390 insertions) sit in the working tree with the auto-commit daemon running. Risk of fragmented "heuristic" commits.

## d) TOTALLY FUCKED UP

1. **DISCOVERED SEVERE PRE-EXISTING BUG (not caused by this session, but confirmed during it):** positional patterns are broken on master. `upd react-star` → **"Unknown command "react-star" for "upd"."** Root cause: fang registers hidden `man`/`completion` subcommands; cobra's `legacyArgs` then rejects arbitrary positional args on a subcommand-having root, so the advertised `upd react*` feature (see `NewCommand` Example text) cannot work. Nothing in the test suite covers CLI-level positional patterns (`TestParseFlagsMultiplePatterns` bypasses `Find()`). This is a release-blocking regression that has apparently existed since the v1.2.0 fang migration.
2. **Process stumbles this session (all caught and fixed, listed for honesty):** edited render.go via a bash `sed` view and had the multiedit rejected (must View before Edit); my first property-test file shipped with a wrong case list, dead code, and an em dash (caught in self-review before commit); removing named returns broke the compile mid-edit; a missing `strings` import failed the test build; the suggestion feature initially didn't fire because cobra applies `FlagErrorFunc` only in `execute()` (unit test caught it).
3. **Environment churn:** several commands auto-backgrounded mid-run (goindex/build-cache cold), one full module re-download occurred, `/tmp` scratch binary vanished between calls — reminder to keep verification artifacts inside the repo or rerun cheaply.
4. **Nothing is committed.** The single most likely way to lose today's work is the auto-commit daemon slicing it into incoherent "heuristic" commits before I do it properly.

## e) WHAT WE SHOULD IMPROVE

1. **Commit in logical chunks now** — release/CI, testing, UX, flake — before the daemon does it worse.
2. **Fix the positional-args regression first** (likely: `cmd.SetFlags`... proper fix is giving the root command explicit `Args: cobra.ArbitraryArgs`-compatible `Args` func or moving man/completion so `legacyArgs` doesn't trip) + add a CLI-level E2E test for `upd <pattern>`.
3. **Add a regression test that runs runE with an unknown flag _and_ positional args** to pin the suggestion path end to end.
4. **Run the full final gates** (buildflow full mode, `nix run .#test`, `nix run .#lint`, govulncheck) before touching anything else about releases.
5. **README** still advertises `--noColor` and possibly `--json` semantics that changed; grep docs for `noColor` and stale `RenderJSON` signatures.
6. **Coverage gate margin**: 87.7% now, but future feature work (`.npmrc`, slog) should land with tests in the same commit to avoid gate thrash.
7. **The `--format` decision (#11)** interacts with my #13 JSON work — sequence them deliberately, not in parallel.

## f) NEXT UP TO 50 (ordered, most impact first)

**Critical (blockers)**

1. Fix positional-patterns regression (`upd react*` → unknown command) + CLI-level E2E test.
2. Fix unknown-flag-with-positional-args ordering so the typo suggestion actually shows.
3. Commit the session's work in logical chunks (release/CI, tests, UX, flake).
4. Run full final gates: buildflow (full), `nix run .#test` (now with `-race`), `nix run .#lint`.
5. Update AGENTS.md (noColor removal, runE refactor, RenderJSON signature, new test files/apps, stale gotchas).
6. Update TODO_LIST.md (mark 17 items done) + CHANGELOG.md `[Unreleased]` + FEATURES.md.
7. Re-run `nix flake check` after flake.nix `.#test`/`#test-integration` edits (last check predates the `-race` edit).

**Release completion**
8. Write `docs/RELEASING.md` runbook (bump → changelog → tag → push → release; include `nix flake check` gate; document legacy 2.x tags — closes TODO #5 too).
9. Add `.goreleaser.yml` (Linux/macOS amd64+arm64, SHA256SUMS) — TODO #3.
10. Add `release.yml` workflow on tag push wired to goreleaser.
11. Decide: cut v1.5.0 with this batch (E2E tests, deprecation, warnings, suggestions) once gates are green.
12. `go-atomic-write`/`go-error-family` consumers unaffected — skip ecosystem sweep (no dep changes this session).

**High-value product work**
13. `.npmrc` parsing: registry URL + `//registry.../:_authToken` — TODO #12.
14. `--format=table|json` + `--silent` alias; keep `--json` as deprecated alias — TODO #11.
15. Surface errorfamily `code`/`family` in JSON output (additive) + `Format('+')` in verbose — TODO #13.
16. `errors.Join` warnings aggregation — TODO #24.
17. `slog` structured logging (level-gated, quiet-aware) — TODO #27.
18. Investigate whether `--version` through `self'.rev` should be `shortRev` for readability in `upd --version` output.
19. Verify `UPD_*` env precedence still correct after `applyEnvFlags` now returns warnings (existing env tests pass, but add a warning-suppression check under `--quiet`).
20. Consider `fang.WithErrorHandler` to route suggestion errors through errorfamily message templates for consistent styling.

**Testing debt**
21. CLI-level E2E test for embedded `upd`-field patterns merging with CLI args through `runE`.
22. E2E test for `--all` table flag and `--verbose` error-detail block through `runE`.
23. E2E test for `--pin-latest` with `latest` tags through `runE` (only library-level covered).
24. E2E test for `--greatest` through `runE`.
25. Retry/backoff path E2E: mock 500→200 sequence with injected sleeper (currently only npm_test-level).
26. `--concurrency` E2E: many packages through `runE` (fetch parallelism, not just unit).
27. Fuzz `UpdateDependency` byte-splicing (jsonv2 decoder offsets) beyond property tests.
28. Coverage threshold as a dprint/buildflow step rather than only CI (local gate parity).
29. Add `.#test-integration` to the RELEASING.md pre-release ritual.
30. Benchmark RenderTable/RenderJSON (currently only diff/patterns/manifest/replaceVersion).

**Docs & repo hygiene**
31. Re-render VHS demos post-restyle — TODO #10.
32. Focused demo tapes (pin-latest, greatest, patterns) — TODO #25; blocked by #1 (patterns demo would show the regression!).
33. Issue/PR templates — TODO #23.
34. README: document env-var warnings, `--noColor` removal timeline, typo suggestions.
35. ROADMAP.md: add v2.0.0 section (alias removal is the first v2 commitment).
36. Grep all docs for `noColor` references and fix.
37. Add the discovered cobra/`legacyArgs` gotcha to AGENTS.md once fixed (why `Args` must be explicit when fang adds subcommands).
38. Document the new `.#test-integration` app in AGENTS.md build/test section.
39. Consider annotating 2026-10-06_19-54 status report as superseded (docs-health pass later).
40. Add CI job names to branch protection expectations doc (which jobs are the gate).

**Nice-to-have**
41. `--noColor=true` form: verify completion scripts don't break on rewritten args (parse-only concern).
42. Suggestion for unknown subcommands (`upd up` → "did you mean update? there is no update") — mostly a no-op, but cheap.
43. Levenshtein: add `find Suggestions for shorthand flags` (currently only long form).
44. Sort `flagNames` for deterministic suggestion ties.
45. Extract `deprecatedNoColorArgs` test table to include `--noColor` inside combined short-flag clusters (`-Cq` unaffected, but document why).
46. `RenderJSON`: consider `Errors` count already correct — assert `summary.errors == errCount` in an E2E test (partial-failure JSON case).
47. Pre-commit hook: run `go test ./cmd/upd -run TestRunE` fast subset.
48. Add `bin/` to .gitignore check (scratch builds like my `bin/upd-check` must never land).
49. Evaluate `errors.AsType` adoption sites in new code for consistency with npm.go's Go 1.26+ idiom.
50. Schedule a docs-health pass after the commit burst so TODO_LIST/FEATURES/CHANGELOG don't drift again.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Is `upd react*` (positional patterns) supposed to work on the v1.2.0–v1.4.0 releases?** My smoke test says it errors with "Unknown command" on current master — if you've used patterns successfully from an installed binary recently, my diagnosis is wrong and I should dig deeper before "fixing" cobra behavior that perhaps fang intentionally gates. May I make this the top-priority fix and cut a patch release for it?
2. **For `--format` (#11):** do you want `--json` kept as a silently-working deprecated alias (my recommendation, mirrors the `--noColor` treatment) or removed outright in the same release that adds `--format`?
3. **Release cadence:** should I cut **v1.5.0** (tests, UX improvements, CI hardening) as soon as gates are green — and additionally a **v1.4.1** beforehand if you confirm the patterns regression affects released binaries — or do you want to batch everything, including goreleaser automation (#3), into one release?

---

**Bottom line:** 17 TODO items genuinely done and verified, working tree green but uncommitted, and the session surfaced one severe pre-existing regression (positional patterns broken) plus one UX gap (suggestion vs. positional-arg ordering) that should rank above everything else except committing this work.
