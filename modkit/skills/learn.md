---
description: Extract generalizable lessons from project postmortems and architecture amendments and propose catalog entries for the registry
---

You are extracting lessons from a completed project or recent incident to propose entries for the cross-project catalog (`learnings/catalog.yaml`). Load context once (§18):
- `~/.modkit/cache/learnings/catalog.yaml` — existing entries (avoid duplicates)
- `~/.modkit/cache/orchestration/composition-rulebook.md` §27 — entry qualification rules

Do NOT re-read files already loaded in this session.

---

## Step 1 — Gather project artifacts

Read from the current project (not the registry):
- `postmortems/*.yaml` — all postmortem files
- `.modkit/runs/*/operate-log.yaml` — Phase 6c findings and action items
- `.modkit/runs/*/self-review.log` — milestones with >1 iteration (same-error retries signal non-obvious bugs)
- `architecture-plan.yaml` — check `deviations:` block for deliberate non-standard choices that worked well

Focus on artifacts that contain `action_items`, high iteration counts, `patterns_to_promote`, or `security_scan` findings.

---

## Step 2 — Identify lesson candidates

For each artifact, ask: "If a new agent read this before starting a similar project, would it change a Phase 2 decision?"

Apply the §27.2 qualification filter:
- ✅ Would have changed an architectural decision if known earlier
- ✅ Non-obvious gotcha that recurs or is easy to miss
- ✅ Generalizable — no project-specific names needed
- ❌ Already in §20–§22 rules — skip (don't duplicate the rulebook)
- ❌ Project-specific context with no reusable principle — skip
- ❌ One-off environment issue — skip

Identify at most 3 candidates. Quality over quantity — 0 candidates is a valid result if nothing qualifies.

---

## Step 3 — Check for duplicates

For each candidate, scan the existing `learnings/catalog.yaml`:
- If an existing entry covers the same root cause and lesson → skip the candidate
- If an existing entry is similar but less specific → propose a refinement to the existing entry instead of a new one
- If the catalog already has >180 entries → propose an entry only if it is clearly higher-value than the lowest-scoring existing entry

---

## Step 4 — Propose catalog entries

For each qualifying, non-duplicate candidate, emit a YAML block:

```yaml
proposed_entry:
  id: "<domain>-<NNN>"          # NNN = next available number in that domain
  category: bug | anti-pattern | integration-gotcha | performance | security | reuse-opportunity
  domains: [<module-names>]
  title: "<≤80-char one-liner>"
  lesson: >
    <Generalized lesson in 2-4 sentences. No project-specific names.
    State what to do (or not do) and why.>
  trigger: "<project signal or condition that makes this lesson relevant>"
  source: "<postmortem ID, amendment ID, or self-review milestone>"
  date_added: "<today's date ISO-8601>"
```

If 0 candidates qualify: emit:
```yaml
learn_result:
  candidates_reviewed: <N>
  entries_proposed: 0
  reason: "No generalizable, non-duplicate lessons identified in this session."
```

---

## Step 5 — Human checkpoint + write

Present all proposed entries (or the zero-result block) to the human.

**🔒 Human checkpoint** — required before any write. Human may:
- Approve all entries → agent appends them to `learnings/catalog.yaml`
- Approve some, reject others → agent appends only approved ones
- Request edits → agent revises and re-presents
- Reject all → session ends with no writes

On approval:
```bash
# Append approved entries to the catalog
# Then commit
git add learnings/catalog.yaml
git commit -m "learnings: add <N> entries from <postmortem-ID or project-name>"
git push origin HEAD:main
```

On rejection: thank the human and stop. Do not re-propose the same entries in the same session.

---

## Token rule

Total output for this skill: ≤300 tokens for the proposed YAML blocks + human-facing explanation. No recap prose after the checkpoint.
