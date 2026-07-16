# PRO/CONTRA: Adopting `cmdguard` into `upd`

**Date:** 2026-07-16
**Decision:** **DO NOT ADOPT**
**Status:** Resolved

---

## The Core Question

`upd` is a **single-command, focused CLI** that bumps versions in `package.json`. It uses Go's stdlib `flag` package with ~100 lines of flag parsing. `cmdguard` is a **full CLI framework** (wrapping Cobra) with DI, lifecycle, subcommands, 16 output formats, styled help, and config files. Should `upd` adopt it?

---

## Measured Impact

| Metric                 | Current (`upd`)          | With `cmdguard`              | Delta     |
| ---------------------- | ------------------------ | ---------------------------- | --------- |
| Direct dependencies    | 4                        | ~14+                         | +10       |
| Total modules (go.sum) | 11                       | ~92                          | **+8.4x** |
| Binary size (stripped) | 7.1 MB                   | ~15–17 MB                    | **~2.2x** |
| CLI structure          | flat, single command     | same (no subcommands needed) | 0 benefit |
| Flag parsing code      | ~100 lines stdlib `flag` | struct tags (less code)      | minor win |

---

## PRO (Arguments FOR)

1. **Type-safe flags via struct tags** — `Config` struct becomes declarative; eliminates the `defineBoolFlag`/`defineStringFlag` boilerplate and manual short/long double-registration.
2. **Auto-generated styled help** — fang + lipgloss produce polished `--help` output without the hand-maintained `PrintUsage()` table in `config.go:201`.
3. **Typo suggestions** — "did you mean --json?" for misspelled flags. Currently absent.
4. **Env var support** — `env:"UPD_REGISTRY"` tag built-in. Currently absent (would require manual `os.LookupEnv` calls).
5. **Shell completion** — free bash/zsh/fish completion scripts.
6. **Zero-panics contract** — all functions return errors. `upd` already does this, but cmdguard enforces it structurally.
7. **Future-proofing** — if `upd` ever grows subcommands (`upd check`, `upd init`, `upd doctor`), cmdguard provides them for free.
8. **Same author ecosystem** — `larsartmann/cmdguard` + `larsartmann/go-error-family` + `larsartmann/go-atomic-write` are philosophically aligned.
9. **Man page generation** — `manpage.GenerateCommand[T](cli)` for free roff output.

---

## CONTRA (Arguments AGAINST)

1. **Dependency bloat — catastrophic**: 11 → 92 modules. Brings in Cobra, pflag, lipgloss, fang, glamour, goldmark, chroma, koanf, samber/do, go-output, go-toml, yaml, and ~60 more. `upd`'s deliberate leanness ("only 4 direct dependencies") is a stated design principle (`AGENTS.md`).
2. **Binary size doubles**: 7.1 MB → ~15–17 MB for a tool that edits one JSON file. Unjustifiable.
3. **Single-command CLI — cmdguard's value is unused**: upd has no subcommands, no DI, no lifecycle management, no config files, no multi-format output. cmdguard's core differentiators (DI with samber/do, `BranchingFlowContext`, `DoctorCommand`, `WithGracefulShutdown`, subcommands) provide **zero benefit**.
4. **Error system conflict**: `upd` already uses `go-error-family` for structured classification (Rejection/Transient/Corruption/Conflict with exit codes). cmdguard brings its own error contract (`ExitCoder`, `NewCommandError`, `NewFlagError`). Two overlapping error systems = confusion.
5. **Signal handling already done**: `upd` has `signal.NotifyContext(SIGINT, SIGTERM)` in `main.go:67`. cmdguard's `WithSignalHandling()` duplicates this.
6. **Output rendering already bespoke**: `render.go` (383 lines) handles the table, colors, verbose mode, error detail blocks. `--json` output is custom (`RenderJSON`). cmdguard's `go-output` 16-format system doesn't replace this — upd's table is domain-specific (old → new version arrows, state badges).
7. **`--dry-run` alias pattern**: upd manually registers `--dry-run` as an alias for `--nop` (`config.go:136`). This kind of domain-specific aliasing works fine with stdlib `flag`.
8. **Embedded `upd` args pattern**: upd reads an `upd` field from `package.json` and prepends those args to CLI patterns (`main.go:41-48`). This is unique domain logic that doesn't map to cmdguard's config file system.
9. **Migration cost**: rewriting `ParseFlags`, `PrintUsage`, `PrintVersion`, `main.go:run()`, flag tests, and re-wiring error handling. ~200 lines of working, tested code replaced by a framework.
10. **Compile time increase**: 92 modules vs 11 measurably slows builds, CI, and `go run` iteration.
11. **`GOEXPERIMENT=jsonv2` interaction**: upd depends on experimental jsonv2. cmdguard uses koanf (go-faster/yaml, mapstructure) for config parsing — a separate JSON/YAML stack that coexists but adds complexity.
12. **Violation of YAGNI**: upd doesn't need 90% of what cmdguard offers. Adding it is a textbook over-engineering case.

---

## Verdict

The cost/benefit ratio is **decisively negative** for `upd`:

| Dimension                  | Assessment                                                        |
| -------------------------- | ----------------------------------------------------------------- |
| **Solves a real problem?** | No — upd's flag parsing works, is tested, and is ~100 lines       |
| **Dependency cost**        | Unacceptable — 8.4x modules, 2.2x binary for zero functional gain |
| **Feature alignment**      | ~10% — only typo suggestions and env vars are genuinely new       |
| **Philosophical fit**      | Violates upd's stated "4 direct deps" leanness principle          |

`cmdguard` is an excellent framework for **multi-command, service-oriented CLIs** that need DI, lifecycle, config files, and rich help. `upd` is none of those things. It's a focused single-purpose tool that edits one file. Using cmdguard here would be like installing Kubernetes to run a cron job.

---

## What `upd` COULD Cherry-Pick (Without `cmdguard`)

If individual features are desired, they can be added standalone — no framework needed:

- **Typo suggestions** → ~30 lines of Levenshtein distance on flag names in `ParseFlags`
- **Env var support** → `os.LookupEnv` calls in `DefaultConfig()` (~10 lines)
- **Better help formatting** → `flag.FlagSet.VisitAll()` to auto-generate help from registered flags (no lipgloss needed)

Total estimated cost: ~50 lines vs 92 transitive modules.
