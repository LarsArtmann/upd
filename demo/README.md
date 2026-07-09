# UPD Demos

Animated demos rendered with [VHS](https://github.com/charmbracelet/vhs).

## Rendering

**Local render (produces GIFs in this directory):**

```
nix run .#demo
```

**Publish to VHS cloud (returns shareable `vhs.charm.sh` URLs):**

```
nix run .#demo -- --publish
```

## Files

| File           | Purpose                                                         |
| -------------- | --------------------------------------------------------------- |
| `demo.tape`    | VHS tape script — types `cat package.json`, `upd -n`, `upd -nP` |
| `package.json` | Demo fixture with outdated deps + a `latest` tag                |

GIF outputs (`*.gif`) are git-ignored — they live in the VHS cloud, not in this repo.
