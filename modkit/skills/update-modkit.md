---
description: Sync the local modkit registry with the latest changes from remote (git pull)
---

You are syncing the local modkit registry to match the remote.

## How the local cache works

`~/.modkit/cache/` is either:
- A **junction/symlink** pointing to the local registry checkout (most common) — updating means `git pull` in that checkout
- A **git clone** — updating means `git pull` in `~/.modkit/cache/` directly

## Steps

1. Resolve the registry path:

```bash
# Check if cache is a symlink/junction — resolve its target
python3 -c "import os; p=os.path.expanduser('~/.modkit/cache'); print(os.path.realpath(p))"
```

If the resolved path differs from `~/.modkit/cache`, it's a junction — use the resolved path.
Otherwise use `~/.modkit/cache` directly.

2. Run git pull in the registry directory:

```bash
git -C <resolved-path> pull --ff-only origin main
```

3. Report the result:
   - If already up to date: tell the user the cache is current, no changes
   - If updated: show the git output (files changed, commits pulled)
   - If it fails (e.g. local uncommitted changes, diverged): show the error and suggest the user resolve it manually

Do not modify any files. Just pull and report.
