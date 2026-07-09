# Domain Language

A **Unified Language** for `upd` — shared across Customer, Product Owner, Developer, and AI.
Inspired by Domain-Driven Design (DDD) Ubiquitous Language.

Every term below should mean the **same thing** to everyone who reads it.
If a word means something different to a developer than to a customer, define it here.

## Glossary

| Term              | Definition                                                                             | Context                                     |
| ----------------- | -------------------------------------------------------------------------------------- | ------------------------------------------- |
| upd               | A Go CLI that upgrades NPM package dependency versions in `package.json`               | The project/product name                    |
| Package file      | The `package.json` being read and potentially modified                                 | The input/output artifact                   |
| Byte-preserving   | Only the version bytes inside a constraint string change; all other formatting is kept | The core guarantee that distinguishes upd   |
| Constraint string | The value portion of a dependency entry (e.g. `"^4.18.0"`, `"latest"`, `">=1.0.0"`)    | What gets classified, compared, and updated |

## Entities

Objects with identity and lifecycle.

| Term        | Definition                                                                       | Context                                                         |
| ----------- | -------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| PackageFile | In-memory representation of `package.json`: raw bytes + xxhash64 fingerprint     | Read once, mutated in place, written atomically                 |
| Spec        | One dependency occurrence in one section of `package.json`                       | The unit of work — classified, fetched, and potentially updated |
| Manifest    | `map[string][]*Spec` — every occurrence of a dependency across all four sections | Built once from PackageFile; drives the engine                  |

## Value Objects

Immutable objects defined by attributes.

| Term        | Definition                                                                        | Context                                                  |
| ----------- | --------------------------------------------------------------------------------- | -------------------------------------------------------- | ------- | ----------------------- | ------------------------------------ |
| Config      | Parsed CLI flags + patterns: file path, concurrency, greatest/nop/pinLatest, etc. | Built once from args; never mutated after parse          |
| Packument   | NPM registry JSON for one package (`dist-tags`, `versions`, ...)                  | Held as raw bytes; queried for version resolution        |
| Fingerprint | xxhash64 hash of the original file bytes captured at read time                    | Guards the atomic write against concurrent modifications |
| State       | The lifecycle stage of a Spec: `todo → check → {skipped                           | kept                                                     | updated | error}`, plus `ignored` | Drives rendering and write decisions |

## State Transitions

The lifecycle of a single `Spec` through the pipeline:

```
ignored          name didn't match any glob pattern (terminal)
    │
todo ──────────── name matched a pattern
    │
    ├── versionRe doesn't match AND not (pinLatest AND latestRe) ──→ skipped (terminal)
    │
check ─────────── version extracted, ready for registry fetch
    │
    ├── registry error ──────────────────────────────────────────→ error (terminal)
    ├── resolved version ≤ current version ──────────────────────→ kept (terminal)
    └── resolved version > current version ──────────────────────→ updated (terminal)
```

Exception: when `IsLatest=true` (bare `"latest"` tag with `--pinLatest`), `shouldUpdate` short-circuits to always update.

## Commands

Actions the system can perform.

| Term             | Definition                                                                 | Context                              |
| ---------------- | -------------------------------------------------------------------------- | ------------------------------------ |
| BuildManifest    | Extracts deps from four sections, applies glob filtering, classifies state | Pipeline step 4-5                    |
| FetchAll         | Concurrently fetches packuments from the NPM registry (semaphore-bounded)  | Pipeline step 6                      |
| ApplyUpdates     | Resolves target versions, compares semver, mutates PackageFile bytes       | Pipeline step 7                      |
| UpdateDependency | Surgically replaces version bytes in raw JSON using jsontext streaming     | Called per-updated-spec              |
| Write            | TOCTOU-safe atomic write: temp file → fsync → fingerprint verify → rename  | Pipeline step 8, only if updates > 0 |

---

> **How to use this file:**
>
> - Keep terms concise — one clear sentence per definition
> - Update when new domain concepts emerge
> - Use these terms consistently in code, docs, and conversations
> - When in doubt about a word's meaning, check here first
