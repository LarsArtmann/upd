# Atomic Writes with TOCTOU Protection

`upd` writes `package.json` through
[go-atomic-write](https://github.com/larsartmann/go-atomic-write) (v0.2.0),
which provides TOCTOU-safe writes via xxhash64 fingerprint verification,
cross-platform file locking, atomic rename, and fsync for crash durability.

## The Problem

During the network-fetch window (which can last several seconds for large
dependency trees), another process — `npm install`, an IDE formatter, a
git hook — may modify `package.json`. If `upd` simply wrote its in-memory
copy back to disk, those external changes would be silently overwritten.

## How It Works

The write path performs these steps in order:

1. **Stage a temp file** — creates `package.json.<random-hex>.tmp` with
   the updated bytes. The random suffix avoids collisions between
   concurrent upd runs.

2. **Fsync the temp file** — calls `file.Sync()` to flush the temp file's
   contents to stable storage. This guarantees the data survives a crash
   that happens after the rename completes.

3. **Acquire an exclusive file lock** — uses cross-platform locking
   (`flock` on Unix, `LockFileEx` on Windows) on the target path to
   serialize the verification + rename window against other processes
   that also respect the lock.

4. **Verify the fingerprint** — re-reads the on-disk `package.json` and
   compares its xxhash64 against the fingerprint captured at read time.
   If they differ, another process modified the file during the fetch
   window.

5. **Atomic rename + directory fsync** — if the fingerprint matches,
   performs `os.Rename(tmp, target)` (atomic on POSIX) and then
   `fsync`s the parent directory so the rename itself is durable.

6. **Abort on mismatch** — if the fingerprint does NOT match, the temp
   file is cleaned up and `ErrConcurrentModification` is returned. The
   original `package.json` is left completely untouched. No `.bak`
   artifacts are left behind.

## Flow Diagram

```
Read package.json
  │
  ├─ capture xxhash64 fingerprint
  │
  ├─ fetch packuments (network, concurrent)
  ├─ apply byte-level version edits in memory
  │
  ▼
Write(path, updatedBytes, fingerprint)
  │
  ├─ write temp file (random suffix)
  ├─ fsync temp file
  │
  ├─ acquire flock on target
  ├─ re-read target from disk
  ├─ compare fingerprint
  │     ├─ match   → atomic rename + fsync dir
  │     └─ mismatch → cleanup temp, return ErrConcurrentModification
  │
  └─ release flock
```
