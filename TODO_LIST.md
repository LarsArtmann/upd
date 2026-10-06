# TODO List

> Short- and mid-term improvement tasks, verified against the actual codebase.
> Harvested from status reports (`docs/status/`), release audits, and code reads.
> De-duplicated. **Open items only** — completed work lives in `CHANGELOG.md`.
> Swept 2026-10-06: 24 of the 28 open items completed; the remainder follows.

---

## Release

| #  | Task                                                                                                                              | Source                     | Notes                                                                                       |
| -- | --------------------------------------------------------------------------------------------------------------------------------- | -------------------------- | ------------------------------------------------------------------------------------------- |
| 1  | Cut the next release (v1.5.0) from current master — patterns fix, `--format`, `.npmrc`, JSON error codes, release automation       | 2026-10-06 sweep           | Gates must be green first; `docs/RELEASING.md` has the ritual.                              |
| 2  | Add signing/provenance to goreleaser artifacts (cosign keyless or checksum signature)                                             | TODO_LIST #3 (partial)     | `SHA256SUMS` exists; signatures not yet generated.                                          |

## Product

| #  | Task                                                                                                   | Source               | Notes                                                                                                     |
| -- | ------------------------------------------------------------------------------------------------------ | -------------------- | --------------------------------------------------------------------------------------------------------- |
| 4  | `.npmrc` scoped registries: honor `@scope:registry=<url>` and per-scope tokens for `@scope/pkg` fetches | TODO_LIST #12 (gap)  | Default registry + tokens work; scope dispatch needs per-package registry resolution in the fetch path.   |

---

## REJECTED (with reasoning)

| #   | Task                                                 | Reason                                                                                                                           |
| --- | ---------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| R1  | Dockerfile                                           | Single static binary makes Docker unnecessary.                                                                                   |
| R2  | `BuildManifest` options struct                       | YAGNI — no external library consumers; positional bool is fine.                                                                  |
| R3  | `Spec.Section` typed enum                            | YAGNI — bare string works; no bugs from it.                                                                                      |
| R4  | `PackageName` branded type                           | YAGNI — adds ceremony without preventing real bugs.                                                                              |
| R5  | Golden file tests vs `rse/upd`                       | Go port has different output format; byte-for-byte parity is artificial.                                                         |
| R6  | `sjson` for writes                                   | Current `jsontext.Decoder` byte-splice approach works and is tested.                                                             |
| R7  | Surface `ErrRegistryUnavailable` in non-fatal path   | Partial failure mixes 404 + 503; exit 1 is correct. Exit 75 reserved for total registry failure.                                 |
| R8  | Additional exit codes (65=EX_DATAERR, 66=EX_NOINPUT) | Only 0, 1, 75 used; adding more adds complexity without clear value.                                                             |
| R9  | `--fail-on-error` flag                               | Resolved — non-zero exit is the default behavior (`ErrPartialFailure`); no flag needed.                                          |
| R10 | `upd update` self-update command                     | Go binaries don't self-update; installs are managed by nix/go install. Same reasoning as the rejected update-notifier.           |
| R11 | Deleting the legacy `2.x` git tags                   | Irreversible, and nothing breaks from their existence once documented (`docs/RELEASING.md`).                                     |
| R12 | `slog` structured logging                            | upd's contract is "render a table or JSON, fail with classified errors". A second logging channel duplicates structured output; retry visibility already exists via `--verbose` error chains. Revisit only if long-running or daemon use cases appear. |
