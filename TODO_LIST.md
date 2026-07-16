# TODO List

> Short- and mid-term improvement tasks, verified against the actual codebase.
> Derived from status reports, research docs, and code audits. De-duplicated.
> Items marked DONE are kept for traceability — remove after each release.

---

## NOT DONE — Medium Priority

| #   | Task                                                          | Source        | Notes                                                                                 |
| --- | ------------------------------------------------------------- | ------------- | ------------------------------------------------------------------------------------- |
| 1   | Add `.npmrc` parsing for custom registry config               | e11, #22      | `--registry` flag covers the primary use case. `.npmrc` would add auth token support. |
| 2   | Add release automation (GoReleaser or tag-based pipeline)     | #9, #13       | No release workflow                                                                   |
| 3   | Add Renovate/Dependabot config                                | #24, f38      | No dependency automation                                                              |
| 4   | Add `nix flake check` to CI                                   | #37           | Not in CI                                                                             |
| 5   | Add coverage threshold to CI (fail if <80%)                   | #23           | Not in CI                                                                             |
| 6   | Add shell completions (bash/zsh/fish)                         | #20, #47, f28 | Not implemented                                                                       |
| 7   | Add man page (`man/upd.1`)                                    | #48, f41      | Not implemented                                                                       |
| 8   | Add property-based tests for `versionRe` and `latestRe` regex | f18           | Not implemented                                                                       |
| 9   | Add Go doc examples with `// Output:` to `doc.go`             | f13           | Example exists but not compile-tested                                                 |
| 10  | Consider `errors.Join` for multi-error aggregation            | err20         | Currently N separate warnings                                                         |
| 11  | Add focused demo tapes (`pin-latest.tape`, `greatest.tape`)   | f11           | Only one tape exists                                                                  |
| 12  | Add integration test (build-tagged) hitting real NPM registry | f12           | All tests use mocks                                                                   |
| 13  | Add issue/PR templates to `.github/`                          | f45           | Not implemented                                                                       |
| 14  | Review all error messages for user-facing quality             | err37, f37    | Not audited                                                                           |
| 15  | Add `slog` structured logging                                 | e15, err28    | No logging stack                                                                      |

## REJECTED (with reasoning)

| #   | Task                                                 | Reason                                                                                           |
| --- | ---------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| R1  | Dockerfile                                           | Single static binary makes Docker unnecessary.                                                   |
| R2  | `BuildManifest` options struct                       | YAGNI — no external library consumers; positional bool is fine.                                  |
| R3  | `Spec.Section` typed enum                            | YAGNI — bare string works; no bugs from it.                                                      |
| R4  | `PackageName` branded type                           | YAGNI — adds ceremony without preventing real bugs.                                              |
| R5  | Golden file tests vs `rse/upd`                       | Go port has different output format; byte-for-byte parity is artificial.                         |
| R6  | `sjson` for writes                                   | Current `jsontext.Decoder` byte-splice approach works and is tested.                             |
| R7  | Surface `ErrRegistryUnavailable` in non-fatal path   | Partial failure mixes 404 + 503; exit 1 is correct. Exit 75 reserved for total registry failure. |
| R8  | Additional exit codes (65=EX_DATAERR, 66=EX_NOINPUT) | Only 0, 1, 75 used; adding more adds complexity without clear value.                             |
| R9  | `--fail-on-error` flag                               | Resolved — non-zero exit is the default behavior; no flag needed.                                |
