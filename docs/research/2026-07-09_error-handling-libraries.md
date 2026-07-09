# Research: Should `upd` adopt go-error-family, its bridge module, and/or samber/oops?

**Date:** 2026-07-09
**Status:** Recommendation produced
**Verdict (TL;DR):** Do **not** adopt any of the three right now. None of them solve a problem `upd` actually has today (no retry loop, no HTTP API, no observability stack). Revisit only if `upd` grows one of those. Full reasoning below.

---

## 1. The three candidates at a glance

|                  | **go-error-family** (root)                                                                                                                                                                | **bridge** (submodule)                                                                                        | **samber/oops**                                                                                                               |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| **What it does** | Classifies errors into behavioral Families (Rejection / Conflict / Transient / Corruption / Infrastructure) that drive exit codes, retry decisions, HTTP status, and user-facing messages | Connects go-error-family + oops: wraps an `oops.OopsError` with a Family, infers Family from oops tags/domain | Enriches errors with stack traces, structured key-value context, trace/span IDs, source fragments, assertions, panic recovery |
| **Kind**         | Classification + boundary protocol                                                                                                                                                        | Integration glue                                                                                              | Enrichment / observability                                                                                                    |
| **Direct deps**  | 0 (stdlib only)                                                                                                                                                                           | go-error-family + oops                                                                                        | 0 (stdlib only)                                                                                                               |
| **Go version**   | 1.26+ (`errors.AsType`)                                                                                                                                                                   | 1.26+                                                                                                         | 1.21+                                                                                                                         |
| **Maturity**     | v0.x, created 2026-05-10, 0 stars, sole author                                                                                                                                            | v0.x experimental submodule                                                                                   | v1 stable, 969 stars, created 2023, active maintenance                                                                        |
| **Strength**     | Answers "what do I _do_ with this error?" — retry? exit code? whose fault?                                                                                                                | Lets you use both libraries together without losing either's metadata                                         | Answers "where did this error come from?" — stack trace, context, trace IDs                                                   |
| **Weakness**     | Young, single-author, unproven in the wild                                                                                                                                                | Pulls in oops as a transitive dep; only useful if you adopt _both_ others                                     | Heavier conceptual model (builder chain, many context fields) than a small CLI needs                                          |

The two libraries are explicitly **complementary, not competing** — go-error-family classifies behavior; oops enriches diagnostics. The `bridge` module exists solely to use both together.

---

## 2. What `upd`'s error handling looks like today

Audited every `errors.` / `fmt.Errorf` call in the repo (74 matches across 11 files).

### Current architecture

- **11 sentinel errors** in `errors.go` (`ErrFileNotFound`, `ErrInvalidJSON`, `ErrConcurrentModification`, `ErrPackageNotFound`, etc.)
- **Wrapping**: `fmt.Errorf("context: %w", err)` everywhere — the idiomatic stdlib pattern
- **Matching**: `errors.Is` only (control-flow checks for `ErrHelp`, `ErrVersion`, `ErrConcurrentModification`). No `errors.As` anywhere.
- **Boundary**: `cmd/upd/main.go` does `fmt.Fprintf(os.Stderr, "\x1b[31mERROR:\x1b[0m %v\n", err)` + `os.Exit(1)` — one exit code for everything
- **No HTTP layer, no logging stack, no retry loop, no database, no observability pipeline**
- **Depguard** in `.golangci.yml` is locked to `$gostd` + exactly 3 allowed imports. Adding any new package requires editing both the `main` and `cmd` allow rules.

### How the 11 sentinels would map to Families (hypothetical)

| Sentinel                             | Natural Family               | Exit code today | Exit code go-error-family would give |
| ------------------------------------ | ---------------------------- | --------------- | ------------------------------------ |
| `ErrFileNotFound`                    | Rejection                    | 1               | 1 (same)                             |
| `ErrInvalidJSON`                     | Corruption                   | 1               | 65 (EX_DATAERR)                      |
| `ErrConcurrentModification`          | Conflict                     | 1               | 1 (same)                             |
| `ErrPackageNotFound`                 | Rejection                    | 1               | 1 (same)                             |
| `ErrVersionParse`                    | Rejection                    | 1               | 1 (same)                             |
| Network / registry failures (npm.go) | Transient                    | 1               | **75 (EX_TEMPFAIL)**                 |
| `ErrHelp` / `ErrVersion`             | _(control flow, not errors)_ | 0               | N/A                                  |

The **only** exit-code change that would be genuinely user-visible and useful: a Transient network failure (NPM registry timeout / 5xx) getting exit code 75 instead of 1, so a CI wrapper script could `&& retry` on 75 but fail-fast on 1.

---

## 3. Pro / Contra analysis

### 3.1 go-error-family (root module)

#### PRO

| #   | Argument                                                                                                                                                                               | Weight for `upd`                                                                          |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------- |
| P1  | **Exit-code differentiation**: Transient registry failures → exit 75 lets CI distinguish "retry me" from "fix your config". This is the single strongest argument.                     | Medium — real but narrow value                                                            |
| P2  | **`HandleError(err)` replaces manual stderr formatting**: `os.Exit(errorfamily.HandleError(err))` is cleaner than the current `fmt.Fprintf + os.Exit(1)`. Saves ~5 lines in `main.go`. | Low — the current code is 3 lines and already clear                                       |
| P3  | **Structured What/Why/Fix/WayOut messages**: richer user-facing errors than raw `err.Error()`.                                                                                         | Low — `upd`'s user is a developer who benefits from the raw error, not a softened message |
| P4  | **Zero dependencies** — root module is stdlib-only, consistent with `upd`'s minimalism.                                                                                                | Medium — doesn't violate the "small dep tree" principle at the transitive level           |
| P5  | **Classification of npm registry responses**: a 404 is Rejection (typo), a 500/timeout is Transient (retry). Currently both are indistinguishable to the caller.                       | Low — `upd` doesn't retry, so the classification has no consumer                          |
| P6  | **Future-proofing**: if `upd` ever adds a `--retry` flag, `IsRetryable(err)` is already there.                                                                                         | Low — YAGNI until the flag exists                                                         |

#### CONTRA

| #   | Argument                                                                                                                                                                                                                                                                                               | Weight                  |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------- |
| C1  | **Depguard allowlist must be edited** (two rules: `main` and `cmd`) for each new import. Not a blocker — just a mechanical step — but it means the dependency choice is a deliberate one, not a freebie.                                                                                               | Low                     |
| C2  | **The library is v0.x, 0 stars, 2 months old, single-author.** No battle-testing, no community, no third-party usage reports. Adopting it in a tool meant to be stable and trustworthy is premature.                                                                                                   | **High**                |
| C3  | **Go 1.26 requirement (`errors.AsType`).** `upd` is on `go 1.26.4` so this is _currently_ satisfied, but it pins the floor higher than the 3 existing deps require, reducing portability for contributors on older toolchains.                                                                         | Low — already satisfied |
| C4  | **Invasive migration for thin value.** To get _any_ classification benefit, all 11 sentinels must either implement `ErrorFamily()` or be registered via `RegisterClassification`. That's 11 code changes + an `init()` block, for a tool whose entire error surface is ~6 wrapping sites in `main.go`. | **High**                |
| C5  | **Not sanctioned by the how-to-golang skill.** The skill's canonical error stack is `cockroachdb/errors + uniflow`. go-error-family is neither blessed nor banned, but it's a lateral move from the recommended path with no clear upside over the recommended stack.                                  | Medium                  |
| C6  | **The `HandleError` pipeline (templates, diagnostics, AI agent) is 95% unused.** `upd` needs none of: message templates, diagnostic rules, HTTP middleware, slog logging, AI debug agent. Adopting the library pulls in conceptual surface area that will never be exercised.                          | Medium                  |

#### Verdict: go-error-family

**Do not adopt.** The single concrete benefit (exit-code 75 for Transient failures) can be achieved in ~10 lines of stdlib code without any dependency (see Alternative below). The library is designed for programs with boundaries (CLI/HTTP/gRPC), retry loops, and observability stacks — `upd` has none of these.

---

### 3.2 samber/oops

#### PRO

| #   | Argument                                                                                                       | Weight for `upd`                                                                                                          |
| --- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| P1  | **Stack traces** for debugging: when `packagejson.go` fails to parse a key, `oops` shows the exact call chain. | Low — the error messages already include `"read dependency key: %w"` style context; the file is 213 lines with clear flow |
| P2  | **Structured context fields** (`.With("package", name).With("section", section)`).                             | Low — `upd` already embeds this context in the error _message string_ via `fmt.Errorf("... %q: %w", name, err)`           |
| P3  | **Mature, stable (v1), 969 stars, active maintenance, zero deps.**                                             | Medium — if any of the three were adopted, this is the safest bet                                                         |
| P4  | **slog integration** — structured logging ready if `upd` adds observability.                                   | Low — no logging stack exists or is planned                                                                               |

#### CONTRA

| #   | Argument                                                                                                                                                                                                          | Weight   |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| C1  | **`upd` is a CLI, not a long-running service.** Stack traces and trace IDs are for Sentry / distributed tracing / log aggregation. A CLI that runs in 2 seconds and exits has no consumer for this enrichment.    | **High** |
| C2  | **The builder-chain style (`oops.In("...").Tags("...").With("...", v).Wrapf(err, "...")`) is heavier than `fmt.Errorf("...: %w", err)`.** For 6 call sites in a focused tool, this adds ceremony without clarity. | **High** |
| C3  | **Same depguard allowlist edit as go-error-family** (C1 above).                                                                                                                                                   | Low      |
| C4  | **`wrapcheck` linter is enabled** — it already enforces wrapping at every boundary. Replacing `fmt.Errorf` with `oops.Wrapf` satisfies the linter equally but adds a dependency to do the same job.               | Medium   |
| C5  | **`errorlint` linter is enabled** — it checks `%w` usage. oops's `Wrapf` uses a custom format that sidesteps this check, potentially masking issues the linter would catch.                                       | Low      |

#### Verdict: samber/oops

**Do not adopt.** oops solves the problem "I have an error in production and I don't know where it came from." `upd` is a local developer tool where the error messages + source file context are already sufficient. Stack traces in a 2-second CLI run are noise, not signal.

---

### 3.3 bridge submodule

#### PRO

| #   | Argument                                                                                                    | Weight                                  |
| --- | ----------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| P1  | **Only module that lets you use classification + enrichment together** without losing metadata from either. | N/A — moot if neither parent is adopted |

#### CONTRA

| #   | Argument                                                                                                                                                      | Weight    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------- |
| C1  | **Only useful if you adopt BOTH go-error-family AND oops.** Since the recommendation is to adopt neither, the bridge has no reason to exist in this codebase. | **Fatal** |
| C2  | **Pulls in two dependencies** (the heaviest option of all three candidates).                                                                                  | **High**  |
| C3  | **v0.x experimental submodule** — the least mature of all three options.                                                                                      | **High**  |

#### Verdict: bridge

**Do not adopt.** It is conditional on adopting both parents, both of which are individually rejected.

---

## 4. Decision matrix

| Criterion                               | go-error-family                  | oops                    | bridge                     | Stay on stdlib       |
| --------------------------------------- | -------------------------------- | ----------------------- | -------------------------- | -------------------- |
| Fits `upd`'s use case (single-pass CLI) | Partial                          | Poor                    | Poor                       | **Perfect**          |
| Respects "right tool for the job"       | No                               | No                      | No (adds 2)                | **Yes**              |
| Maturity / risk                         | Low (v0.x, new)                  | High (v1, 969★)         | Lowest (v0.x experimental) | **Highest (stdlib)** |
| Concrete benefit today                  | Exit code 75                     | Stack traces            | None standalone            | **None needed**      |
| Migration cost                          | 11 sentinels + init() + depguard | 6 call sites + depguard | Both migrations + depguard | **Zero**             |
| Sanctioned by how-to-golang skill       | No (lateral)                     | No (lateral)            | No                         | **Yes (status quo)** |

---

## 5. Recommendation

### Primary: Stay on the stdlib

`upd`'s error handling is already idiomatic Go: sentinel errors, `%w` wrapping, `errors.Is`. The tool is too small and too focused to benefit from a classification protocol or an enrichment framework — none of the three libraries solve a problem the tool currently has.

### If the one real benefit (exit code 75) is still desired

Implement it in ~10 lines with zero dependencies:

```go
// In cmd/upd/main.go
func exitCode(err error) int {
    if err == nil {
        return 0
    }
    if errors.Is(err, upd.ErrConcurrentModification) {
        return 1
    }
    // Registry/network failures are transient — CI can retry
    var netErr *net.OpError
    if errors.As(err, &netErr) {
        return 75 // EX_TEMPFAIL
    }
    return 1
}
```

This captures 90% of go-error-family's value for `upd` at zero dependency cost.

### When to revisit

Revisit adoption **only if** `upd` grows into one of these:

| Trigger                                                  | Candidate                                                           |
| -------------------------------------------------------- | ------------------------------------------------------------------- |
| `--retry` flag for transient registry failures           | go-error-family (Family-driven retry)                               |
| Structured logging / Sentry integration                  | oops (enrichment + stack traces)                                    |
| HTTP API or daemon mode                                  | go-error-family (HTTPStatus middleware) + oops (trace IDs) + bridge |
| Exit-code-sensitive CI integration is requested by users | go-error-family OR the 10-line stdlib snippet above                 |

Until one of these materializes, **the stdlib is the right tool for this job.**

---

## 6. Sources

- **go-error-family README**: `github.com/LarsArtmann/go-error-family` (master branch, fetched 2026-07-09)
- **go-error-family SKILL.md**: full architecture + API reference (master branch)
- **samber/oops README**: `github.com/samber/oops` (main branch, fetched 2026-07-09) — 969 stars, MIT, v1 stable
- **upd codebase audit**: `errors.go`, `cmd/upd/main.go`, `npm.go`, `packagejson.go`, `.golangci.yml`, `go.mod`
- **how-to-golang skill**: banned-libraries.md (error handling section: `cockroachdb/errors + uniflow` is canonical), key-patterns.md
