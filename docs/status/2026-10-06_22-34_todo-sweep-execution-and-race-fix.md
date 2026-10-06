# Status — TODO sweep execution: regression fix, UX overhaul, .npmrc, release automation, and a found data race

- **Date:** 2026-10-06 22:34
- **Scope:** Executed the deferred execution plan from the 21:14 report: fixed the severe
  positional-patterns regression (TDD), completed 7 TODO items (#11, #12, #13, #23, #24, #3/#4/#5,
  #10/#25), rejected 1 (#27), fixed a pre-existing data race the new `-race` gate exposed, updated
  all project docs, and committed everything in 9 logical commits. All final gates green.
  **Nothing pushed; no release cut (deliberate — needs your call).**

---

## Commits (oldest → newest)

| Commit    | Subject                                                                                                          |
| --------- | ---------------------------------------------------------------------------------------------------------------- |
| `8f8d6fd` | Fix broken positional patterns, harden CI, and expand test coverage (the 21:14 session's sweep, committed first) |
| `827b163` | Expose machine-readable error codes in --json output (#13)                                                       |
| `2ea7868` | Replace --json with --format and add a --silent alias (#11)                                                      |
| `0223a2a` | Read registry URL and auth tokens from .npmrc (#12)                                                              |
| `3696a33` | Carry every package failure in the partial-failure error (#24)                                                   |
| `cced077` | Add issue and pull request templates (#23)                                                                       |
| `cfb5cfd` | Add release automation and a release runbook (#3/#4/#5)                                                          |
| `536360b` | Update docs for the format flag, .npmrc support, and release flow (incl. demo tapes #10/#25)                     |
| `1cacb3e` | Fix data race in the progress bar under concurrent fetches (found by the new -race gate)                         |

---

## a) FULLY DONE

| Item                        | What                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        | Verification                                                                                                                                                                                                                                                |
| --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Patterns regression**     | Root cause: fang's hidden `man`/`completion` subcommands make cobra's `legacyArgs` reject positionals (`unknown command "react*"`). Fix: `Args: cobra.ArbitraryArgs` on the root command (`config.go:153`).                                                                                                                                                                                                                                                                                 | TDD: 2 new e2e tests written first, confirmed failing with the exact diagnosed error, green after the one-line fix; live smoke: `upd definitely-not-a-real-pkg-xyz` now fetches (404 → error state → exit 1), `man`/`completion`/`--version` routing intact |
| **Suggestion ordering**     | Same root fix: with `Args` set, `Find()` no longer errors, flag parsing runs, `FlagErrorFunc` fires even with positional args present (`upd --jso -f pj.json` → "Did you mean --format=json?").                                                                                                                                                                                                                                                                                             | `TestRunEUnknownFlagWithPositionalArgSuggestsFlag`                                                                                                                                                                                                          |
| Lint debt from sweep commit | All 4 findings on session-new code fixed properly (no nolint): makezero ×3 (append-style builders), varnamelen (`ar, br` → `fromRunes, toRunes`), intrange ×2                                                                                                                                                                                                                                                                                                                               | `golangci-lint run ./...` → 0 issues                                                                                                                                                                                                                        |
| TODO **#13**                | JSON error entries now carry `code` + `family` via `errorfamily.Code`/`Classify` (`render.go:jsonError`). Additive + `omitempty`.                                                                                                                                                                                                                                                                                                                                                           | `TestRenderJSONIncludesErrors` extended (asserts `registry.package_not_found` / `rejection`)                                                                                                                                                                |
| TODO **#11**                | `--format=table\|json` replaces `--json` (invalid values rejected early with `ErrInvalidFormat`, new 14th sentinel); `--silent`/`-s` registered alias for `--quiet`; `--json` + `UPD_JSON` still work via CLI-layer rewrite + deprecation warning (`rewriteDeprecatedArgs`); removed-flag typos suggest replacements (`--jso` → `--format=json (--json is deprecated)`).                                                                                                                    | 12 new/updated tests (format parse/reject, silent alias, env precedence `UPD_FORMAT` wins over `UPD_JSON`, deprecation warning e2e, man-page `--json` count = 0); live help smoke shows `--format` + `-s --silent`, no `--json`                             |
| TODO **#12**                | `.npmrc` support (`npmrc.go`): registry URL + `//host/:_authToken` bearer tokens; home-then-project precedence (project wins); `--registry`/`UPD_REGISTRY` beats file values; `Config.RegistryToken` → `Authorization: Bearer` in `RegistryClient`. Token matching: scheme-less, trailing-slash tolerant, longest-prefix wins, path-boundary guard (`/team` ≠ `/team-extra`). Legacy `_auth` rejected with a warning. Token values never appear in output/errors.                           | 9 unit tests (`npmrc_test.go`) + e2e `TestRunEEndToEndNpmrcRegistryAndAuth` (asserts registry URL from `.npmrc` AND the `Authorization` header on the wire)                                                                                                 |
| TODO **#24**                | Partial failure now returns `errors.Join(ErrPartialFailure.WithContextf(...), <every Spec.Err>)` (`aggregateFailures` in `cmd/upd/main.go`). `errors.Is`/`As` can inspect individual failures; rendered output unchanged; exit code still 1.                                                                                                                                                                                                                                                | e2e asserts `errors.Is(err, ErrPackageNotFound)` through the join + exit 1; live smoke shows summary + detail block                                                                                                                                         |
| TODO **#23**                | `.github/ISSUE_TEMPLATE/bug_report.yml` + `feature_request.yml` (YAML-validated) + `PULL_REQUEST_TEMPLATE.md` (why / changes / verification + docs checklist).                                                                                                                                                                                                                                                                                                                              | `yaml.safe_load` both forms                                                                                                                                                                                                                                 |
| TODO **#3** (builds half)   | `.goreleaser.yml`: static builds, Linux+macOS, amd64+arm64, `GOEXPERIMENT=jsonv2` in build env, `ProgramVersion` injected from tag, tar.gz archives + `SHA256SUMS`, `release.mode: append`.                                                                                                                                                                                                                                                                                                 | `goreleaser check` → validated; `goreleaser release --snapshot` → 4 archives + checksums in 26s; extracted Linux binary runs and reports version                                                                                                            |
| TODO **#3** (workflow half) | `.github/workflows/release.yml` on `v*` tag push; goreleaser-action pinned to **verified** v6.0.0 SHA (`286f3b1...` via `git ls-remote`, not guessed); checkout/setup-go SHAs copied from ci.yml policy.                                                                                                                                                                                                                                                                                    | YAML follows repo SHA-pinning convention                                                                                                                                                                                                                    |
| TODO **#4 + #5**            | `docs/RELEASING.md`: versioning scheme, pre-flight gate battery, exact cut commands, automation handoff, verification steps, **legacy 2.x tags** section (closes #5), rollback/module-proxy-poisoning guidance.                                                                                                                                                                                                                                                                             | Written this session                                                                                                                                                                                                                                        |
| TODO **#10 + #25**          | New tapes `patterns.tape` / `greatest.tape` / `pin-latest.tape` + two fixtures; all 4 rendered via `nix run .#demo` **and published** to vhs.charm.sh; README now embeds the new main demo and links the three focused ones. The patterns demo exercises the fixed regression against the live registry.                                                                                                                                                                                    | Render job output shows 4 published URLs; GIFs exist locally (git-ignored)                                                                                                                                                                                  |
| TODO **#27**                | Rejected with reasoning (R12 in TODO_LIST): a second slog channel duplicates the table/JSON/errorfamily contract; retry visibility already exists via `--verbose`.                                                                                                                                                                                                                                                                                                                          | TODO_LIST REJECTED section                                                                                                                                                                                                                                  |
| Docs refresh                | README (usage line, flag table, env table, `.npmrc` section, demo links), FEATURES.md (6 rows: noColor, format, env vars, man-page un-yellowed, `.npmrc` PLANNED→FULLY, JSON), AGENTS.md (runE pipeline, step-3 `.npmrc` stage, 4 new gotchas: cobra-Args trap, deprecated-alias mechanism, `.npmrc` security, release automation; sentinel count 13→14; `.#test-integration` in build section), CHANGELOG `[Unreleased]` (full Added/Fixed/Changed), TODO_LIST rewritten (28→4 open items) | Manual review                                                                                                                                                                                                                                               |
| **Data race fix**           | `ProgressReporter` render/Finish now mutex-guarded — concurrent fetch goroutines raced on the writer (found by the new `-race` gate; **pre-existing**, not from this session).                                                                                                                                                                                                                                                                                                              | `go test -race` both packages green; `nix run .#test` (which includes `-race`) fully green after fix                                                                                                                                                        |

**Final gate battery (all green, in order):** `GOEXPERIMENT=jsonv2 GOWORK=off go test . ./cmd/upd -race` → ok · `golangci-lint run ./...` → 0 issues · `nix flake check` → all checks passed · `govulncheck` → No vulnerabilities found · `nix run .#test` (with `-race`) → PASS.

---

## b) PARTIALLY DONE

1. **TODO #3 signatures** — the TODO's original wording was "`SHA256SUMS` + signatures". Checksums ship; signatures/provenance (cosign keyless) do **not**. Logged as TODO_LIST "Release #2".
2. **TODO #12 scoped registries** — `@scope:registry=<url>` and per-scope tokens are **not** honored (default registry + tokens only). Documented in README + TODO_LIST "Product #4". Needs per-package registry resolution in the fetch path — bigger than a parser tweak.
3. **Demo content verification** — I verified the render/publish _process_ (4 URLs, exit 0, files exist) but did not visually inspect each GIF's frames (e.g. whether `greatest.gif` shows any actual update vs an "all up-to-date" box — depends on live registry state at render time).
4. **The release itself** — everything for v1.5.0 is staged (gates green, changelog section drafted, runbook written), but tags are irreversible, so the cut waits for your decision (see g.1).

## c) NOT STARTED

- **Cutting v1.5.0** (and/or a v1.4.1 for the patterns fix — see g.1).
- **Pushing**: the 9 commits sit on local `master`. I never push without your say-so.
- **goreleaser signing/provenance** (cosign) — see b.1.
- **Scoped registry support** — see b.2.
- **Untouched leftovers from the 21:14 "next 50" list**: additional runE-level E2E coverage (`--all`, `--verbose` block, `--greatest`, `--pin-latest`, embedded `upd`-field merge, retry/backoff with injected sleeper, `--concurrency`), fuzzing `UpdateDependency` byte-splicing, coverage gate as a local buildflow step, RenderTable/RenderJSON benchmarks, ROADMAP v2.0.0 section, branch-protection doc, annotating superseded status reports, warning-suppression-under-`--quiet` check.

## d) TOTALLY FUCKED UP

1. **Gate sequencing: I committed before running the complete battery.** The sweep commit (`8f8d6fd`) landed after `go test` + lint were green, but `nix run .#test` (with the new `-race`) ran later — and it caught a **real pre-existing data race** in the progress bar. The gate did its job; my ordering was wrong. A race could have been _introduced_ by that commit and I'd have shipped it in the first commit. The fix is a separate honest commit (`1cacb3e`), but the discipline failure is real: **full battery first, then commit.**
2. **Flawed gate command design**: `go test ... | tail -3 && golangci-lint ... && echo ALL_GREEN` — pipelines mask exit codes (`tail` succeeds even when tests fail), so "ALL_GREEN" printed next to a FAIL once this session. I caught it by reading the output, but the command lied by construction. Gates must use `set -o pipefail` or run unpiped.
3. **`sed -i` on a repo file** (updating suggestion assertions) — broke the edit tool's read-state twice afterward ("file modified since read"), costing re-read round trips. Violates the read-before-edit discipline for the sake of saving one call.
4. **Environment friction (self-inflicted, minor)**: `go run` from `/tmp` without module context failed once; `govulncheck` not on PATH cost a dead background job before rerunning via `go run`; one verification call was interrupted mid-flight and had to be retried.
5. **Untested code path shipped**: the `TokenFor` path-boundary guard (`/team` must not match `/team-extra`, `npmrc.go`) has **no test** — I wrote the guard while fixing the trailing-slash failures but never added the adversarial case to the test table. Currently correct by inspection; that's not verification. (Now TODO_LIST-trackable; see f.6.)

## e) WHAT WE SHOULD IMPROVE

1. **Gate order is law**: full battery (`test -race` → lint → `nix flake check` → govulncheck) BEFORE any commit, always. The first commit of a session is the most likely to be under-verified.
2. **Gates must propagate exit codes**: no `| tail` without `set -o pipefail`. A green-looking pipeline that fails silently is worse than a red one.
3. **No `sed -i` in tracked files** — use the edit tools; their state tracking exists for a reason.
4. **Every guard clause written during a bug fix gets its adversarial test in the same commit** (the `/team-extra` lesson).
5. **Verify rendered artifacts, not just exit codes** — GIFs, release binaries, and published pages deserve at least one content-level assertion each.
6. **release.yml has no test gate** — a tag pushed on a red master would release broken binaries. RELEASING.md documents the rule but nothing enforces it (see f.4).
7. **Pin the goreleaser version in the workflow** — local validation used the nix-installed goreleaser; CI uses whatever the action defaults to. Version drift between "validated locally" and "runs in CI" is a future snapshot-build failure.
8. **Race-detector value proven again**: the `-race` local gate (TODO #28, done last session) caught a real production-facing bug on its first full run. Keep it in the pre-flight ritual and in RELEASING.md's gate list (it's there).

## f) NEXT UP TO 50 (ordered, most impact first)

**Immediate**

1. Decide + cut the release: **v1.5.0** (everything batched) vs **v1.4.1** (patterns fix only) + **v1.5.0** after — runbook is ready either way.
2. Push the 9 local commits to `origin/master` (user action; I don't push unprompted).
3. Pin the goreleaser version in `release.yml` (e.g. `version: v2.x` in the action's `with:`).
4. Add a test+vet job inside `release.yml` before the goreleaser step (protects against tags cut on red master).
5. Cut + verify the release: `gh release view`, install a prebuilt binary, check `--version`, README install instructions for binaries.
6. Add the `/team` vs `/team-extra` boundary test to `npmrc_test.go` (untested guard, see d.5).
7. E2E: `--registry` flag beats `.npmrc` registry (precedence asserted through `runE`, not just unit level).
8. E2E: env-var warnings (incl. `UPD_JSON` deprecation) are suppressed under `--quiet`.
9. E2E: `--all` table flag and `--verbose` error-detail block through `runE`.
10. E2E: `--greatest` and `--pin-latest` through `runE` (library-level covered only).

**Product**
11. Scoped registries: `@scope:registry` + per-scope tokens (fetch-path registry resolution per package).
12. goreleaser artifact signing/provenance (cosign keyless; needs only GitHub OIDC).
13. E2E: embedded `upd`-field patterns merging with CLI args through `runE`.
14. E2E: retry/backoff path with injected sleeper through `runE` (500→200 sequence).
15. E2E: `--concurrency` parallelism through `runE` (many packages).
16. Fuzz `UpdateDependency` byte-splicing beyond property tests.
17. Assert `summary.errors == errCount` in a partial-failure JSON e2e (previous list #46).
18. `--version` output: decide whether the nix `self'.rev` should render `shortRev` for readability (previous list #18).
19. `fang.WithErrorHandler`: route suggestion/parse errors through errorfamily message templates for consistent styling (previous list #20).

**Testing debt**
20. Fuzz/property-test the `.npmrc` parser (malformed lines, duplicate keys, CRLF line endings).
21. Duplicate `registry=` keys in `.npmrc` should warn (currently last wins silently).
22. Man-page test asserting **zero** deprecated aliases (`noColor`, bare `--json`) — regression-proof the mango leak fix.
23. Completion-script test asserting deprecated aliases are not offered.
24. Concurrency=1 regression test asserting progress-bar output determinism (mutex behavior pin).
25. Coverage gate as a local buildflow step (parity with CI's 80% gate).
26. Benchmark `RenderTable`/`RenderJSON` (only diff/patterns/manifest/replaceVersion covered).
27. Shorthand-flag typo suggestions (`-C` cluster typos; currently long form only).
28. Deterministic suggestion tie-breaking (sort `flagNames`).
29. Unknown-subcommand suggestion (`upd up` → "did you mean…?") — cheap.
30. Scoped-package live integration case (`@types/node`) in `integration_registry_test.go`.
31. `--noColor=true` inside combined short-flag clusters: document why `-Cq` works but clusters don't rewrite (previous list #45).
32. `errors.AsType` adoption audit in the new code paths for idiom consistency (previous list #49).
33. Pre-commit hook running the fast `TestRunE` subset (previous list #47).
34. Confirm `bin/` scratch-build gitignore coverage (previous list #48).

**Docs & hygiene**
35. ROADMAP.md: add the v2.0.0 section (alias removal is the first v2 commitment).
36. README: deprecation timeline table (`--noColor`, `--json`, `UPD_JSON` → removed in v2).
37. Branch-protection expectations doc: which CI jobs gate `master` (previous list #40).
38. Annotate superseded status reports (2026-10-06_19-54 and 21-14 are now stale in parts).
39. RELEASING.md: add "re-render + publish demos" as a release step (demos show the CLI; they go stale on restyle).
40. RELEASING.md: add `nix run .#test-integration` already present — verify pre-flight list matches the actual gate battery one-to-one after today's `-race` addition.
41. Schedule the next docs-health pass post-release (FEATURES/TODO_LIST/CHANGELOG drift).
42. Check whether the JS-era `2.x` tags deserve a README footnote (RELEASING.md covers it; user-facing README may not need it).
43. Decide whether `metadata.json`/`artifacts.json` from goreleaser should be committed or ignored long-term (currently only `dist/` is ignored wholesale — fine, just confirm).
44. Dependabot config review: ensure major bumps (gobwas/glob-style API renames) can't silently merge now that the build gate is strict.
45. Consider `changelog.disable: true` in goreleaser — already set; document _why_ in RELEASING.md (CHANGELOG.md is hand-curated).
46. Verify `upd --version` parity: goreleaser-injected tag version vs nix `rev` vs `go install` (document the matrix in RELEASING.md).
47. Demo tapes: add a `--format=json` focused tape (CI-integration story sells itself).
48. Issue template: add a "registry type" dropdown (public / private+token / proxy) to speed triage of `.npmrc` issues.
49. Evaluate `npmrc.go` warnings dedup (same malformed line in user + project file warns twice — acceptable, but confirm desired).
50. Celebrate + schedule the next sweep; the backlog is now 4 items, which is the point.

## g) QUESTIONS I CANNOT ANSWER MYSELF

1. **Release shape**: cut **v1.5.0 now** with everything (patterns fix + `--format` + `.npmrc` + release automation), or cut **v1.4.1 first** (patterns fix cherry-picked, since released v1.2.0–v1.4.0 binaries have the broken `react*` behavior) and v1.5.0 after? My recommendation: single v1.5.0 — the regression fix ships 3 days after v1.4.0, and prebuilt binaries didn't exist before this release anyway.
2. **Signing**: do you want **cosign keyless provenance/signatures** on the goreleaser artifacts in `release.yml` now (needs only GitHub OIDC, no secrets to manage), or is `SHA256SUMS` enough for v1.5.0 and signing lands later?
3. **Release workflow gate**: should `release.yml` run the full test suite (with `-race`) before goreleaser — slower releases, but a tag on a red master can't ship broken binaries — or stay fast and rely on the documented "never tag while CI is red" rule?

---

**Bottom line:** The patterns regression that made the flagship feature unusable is fixed and pinned, 7 more TODO items shipped, a pre-existing data race is dead, release automation exists and was dry-run-verified end to end, and the backlog is down from 28 to 4 items. Everything is committed (9 commits), gated green, and waiting on exactly one decision: the release.
