# Archived status & research reports

Closed history. Every file here was fully resolved inline (each numbered item
carries a `~~strike~~` verdict: done / Won't implement / NOT-DO / superseded)
before being moved. Never reopen items from these files — their open work was
routed to `TODO_LIST.md` / `ROADMAP.md` at annotation time. Reading archived
files during HARVEST/VERIFY/AUDIT burns context and tempts re-opening closed
work.

## 2026-10-06 sweep manifest (docs-health, all `2026-0*` files)

Every remaining unstruck item in each file was verified against the code and
resolved inline before archiving.

| File                                                          | Classification | Deciding reason                                                                                                                           |
| ------------------------------------------------------------- | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `2026-07-09_07-16_pinlatest-migration-and-review.md`          | ARCHIVE        | All 50 items + 2 questions resolved (feature shipped `81d8c44`; test gaps routed to TODO_LIST #5)                                         |
| `2026-07-09_09-08_jsonv2-vhs-migration-and-cleanup.md`        | ARCHIVE        | json/v2 migration shipped; demo infra rendered/published; all P0–P3 items resolved or rejected                                            |
| `2026-07-09_14-52_error-handling-overhaul.md`                 | ARCHIVE        | Overhaul shipped and superseded by the `go-error-family` adoption; all next-steps resolved (f1–f30)                                       |
| `2026-07-09_17-49_todo-list-mass-implementation.md`           | ARCHIVE        | D24–D46 all landed; f1–f50 resolved (most done at `e64d3a7`/`8f8d6fd`; rejects documented)                                                |
| `2026-07-09_18-55_fake-clock-ci-fixes-test-gaps.md`           | ARCHIVE        | Fake clock shipped (`sleeper`); all test-gap items done or routed (TODO_LIST #6/#7)                                                       |
| `2026-07-15_23-30_quality-scan-fixes-partial.md`              | ARCHIVE        | Superseded by the 23:57 round-2 triage; its f-list fully resolved                                                                         |
| `2026-07-15_23-57_quality-scan-round2-self-review.md`         | ARCHIVE        | 128-issue triage documented in AGENTS.md; f1–f50 resolved (usage errors exit 1 as of 2026-10-06)                                          |
| `2026-07-16_00-50_errorfamily-adoption-complete.md`           | ARCHIVE        | Adoption shipped (`db891d0`, `3cd313e`); JSON codes at `827b163`; retryableError decision closed (option a)                               |
| `2026-07-16_05-30_fang-cobra-cli-migration.md`                | ARCHIVE        | Migration + follow-up shipped (`81d8c44`, `ea6493d`, `2ea7868`); the flagged exit-code risk was real and is now fixed + regression-tested |
| `2026-07-16_07-01_fang-followup-color-env-tests.md`           | ARCHIVE        | Follow-up shipped; all 50 f-items resolved (deprecation, env warnings, `--format`, `--silent` at `2ea7868`)                               |
| `2026-07-26_22-24_v1.2.0-release-cut-and-gaps.md`             | ARCHIVE        | Release published (Appendix); automation gap closed at `cfb5cfd`; version truth resolved at `d43f460`                                     |
| `2026-09-25_20-36_docs-health-audit-and-glob-build-fix.md`    | ARCHIVE        | Build break fixed (v1.4.0); its own open items (v1.3.0 release, go.work, CI window) resolved 2026-10-06                                   |
| `../research/archived/2026-07-09_error-handling-libraries.md` | ARCHIVE        | Verdict overturned 2026-07-16 (`go-error-family` adopted); oops/bridge rejections stand; analysis kept for context                        |
| `../research/archived/2026-07-09_superb-error-handling.md`    | ARCHIVE        | Nine-gap design fully implemented; library question resolved; non-change #1 superseded by `ErrPartialFailure`                             |

## Earlier archives (2026-09-25 pass)

| File                                              | Classification | Deciding reason                                                         |
| ------------------------------------------------- | -------------- | ----------------------------------------------------------------------- |
| `2026-06-17_18-03_go-port-complete.md`            | ARCHIVE        | Go port complete; every item struck (57 markers)                        |
| `2026-06-17_18-26_round2-quality-and-devops.md`   | ARCHIVE        | Quality/DevOps round resolved; every item struck (59 markers)           |
| `2026-06-28_04-14_post-atomic-write-upgrade.html` | ARCHIVE        | HTML snapshot; resolutions marked with `<del>`/`<s>` (HTML has no `~~`) |
