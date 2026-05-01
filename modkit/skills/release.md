---
description: Bump semver, generate CHANGELOG from Conventional Commits, create release tag, and trigger production deploy
---

You are creating a new release. Load context once (§18):
- `~/.modkit/cache/orchestration/composition-rulebook.md` §24 (release management)
- The project's `slo.yaml` (for the pre-release checklist)

Do NOT re-read files already loaded in this session.

---

## Step 1 — Pre-release checklist (§24.5)

Verify all of these before proceeding. STOP and report if any fail:

```bash
# 1. All Phase 4 critical flows passing on staging?
# (Ask user to confirm or run smoke tests against staging URL)

# 2. No open P1/P2 incidents?
# (Check postmortems/ directory for any unresolved action_items)

# 3. Secrets not about to expire?
# (Read security-log.yaml — flag any secret due within 14 days)

# 4. CI passing on main?
git log --oneline -5   # last 5 commits
# Ask user to confirm CI is green
```

If any check fails → report the specific blocker and wait for human resolution before continuing.

---

## Step 2 — Determine version bump (§24.2)

Read commits since the last tag:
```bash
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

Apply semver rules:
- Any `feat!:` or `BREAKING CHANGE:` footer → MAJOR bump
- Any `feat:` (without `!`) → MINOR bump
- Only `fix:`, `docs:`, `chore:`, etc. → PATCH bump

If there are non-Conventional Commits in the range → STOP. Ask the user to reword them before continuing. Non-conventional commits block changelog generation.

Report the proposed version:
```
Current: v1.2.3
Proposed: v1.3.0  (reason: feat: found in range)
```

**🔒 Human Checkpoint** — confirm the proposed version before tagging.

---

## Step 3 — Generate CHANGELOG (§24.2, §22 — don't reinvent)

Use the appropriate tool for the runtime:

**Go projects:**
```bash
# Install git-cliff if not present
which git-cliff || go install github.com/orhun/git-cliff@latest

# Generate CHANGELOG for this release
git-cliff $(git describe --tags --abbrev=0)..HEAD --tag v<NEW_VERSION> --prepend CHANGELOG.md
```

**Bun/TS projects (changesets):**
```bash
# If changesets is not set up, use git-cliff as fallback
npx git-cliff $(git describe --tags --abbrev=0)..HEAD --tag v<NEW_VERSION> --prepend CHANGELOG.md
```

Review the generated CHANGELOG section. Present it to the user for review.

**🔒 Human Checkpoint** — confirm CHANGELOG content before tagging.

---

## Step 4 — Create the release

On approval:
```bash
# Commit the CHANGELOG
git add CHANGELOG.md
git commit -m "chore: release v<NEW_VERSION>"

# Tag
git tag v<NEW_VERSION>
git push origin main
git push origin v<NEW_VERSION>
```

The version tag triggers the `deploy-production.yaml` CI workflow (or equivalent for vercel/railway impls).

---

## Step 5 — Verify deploy

After pushing the tag:
1. Monitor CI status (ask user to confirm deploy workflow started)
2. Once deploy completes: `curl -sf https://api.<project>.com/health` (or ask user to verify)
3. Monitor error tracking dashboard for 15 minutes post-deploy

If deploy fails → follow §24.3 rollback protocol:
```bash
git revert <release-commit>
git push origin main
# Then re-run the previous deploy or rollback via platform CLI
```

---

## Output

Post a compact release summary to chat:
```yaml
release:
  version: v<NEW_VERSION>
  bump_reason: "<feat/fix/breaking>"
  commits_in_range: <N>
  changelog_lines: <N>
  tag: v<NEW_VERSION>
  deploy: triggered | failed
```

No narrative prose. Post the YAML block and stop.
