---
description: Run Phase 6 Day-2 operations — SLO setup, architecture evolution, and weekly maintenance checks
---

You are running Phase 6 of the modkit playbook. Load context once (§18):
- `~/.modkit/cache/orchestration/playbook.md` §Phase 6
- `~/.modkit/cache/orchestration/composition-rulebook.md` §23–§26
- The project's `project-brief.yaml` (for project signals)

Do NOT re-read files already loaded in this session.

---

## Determine which sub-phase to run

Check the project's `.modkit/runs/` directory for the latest `operate-log.yaml`. Based on its contents:

| Condition | Run |
|-----------|-----|
| No `slo.yaml` in repo root | Phase 6a (Instrument) |
| `slo.yaml` exists AND a trigger is present (SLO breach, user reports finding, `patterns_deferred` threshold hit) | Phase 6b (Iterate) |
| No trigger AND it has been ≥7 days since last 6c run (or no 6c log exists) | Phase 6c (Maintain) |
| No trigger AND last 6c run was <7 days ago | Report status and wait |

If the user explicitly requests a sub-phase (e.g. "run maintenance checks"), skip the condition check.

---

## Phase 6a — Instrument

Run once after first production deploy.

1. Read `project-brief.yaml` for project signals (§22.3): scale, user type, regulatory context
2. Apply §23.1 SLO defaults adjusted for project signals
3. Emit `slo.yaml` in the repo root (schema in Playbook §Phase 6a)
4. Emit `alerts.yaml` derived from the SLOs:
   - Page: SLO breach sustained >5 min, service down, error spike >10×
   - Warn: budget burn >50% in 1h
5. Emit `data-governance.yaml` (§26.1) using entities from the schema migrations
6. Emit `threat-model.yaml` (§25.4) — one entry per public API surface
7. Emit `pii-inventory` block to `architecture-plan.yaml` (§26.3)

Present all artifacts for review. **🔒 Human Checkpoint** — wait for approval before Phase 6b begins.

Append to `operate-log.yaml`:
```yaml
- cycle: instrument
  phase: "6a"
  date: <ISO-8601>
  slo_yaml_committed: true
  action_required: false
  tokens_used: <count>
```

---

## Phase 6b — Iterate

Run when a measured signal triggers an evolution.

1. Collect the signal:
   - SLO breach from `alerts.yaml` thresholds
   - `patterns_deferred` entry whose `trigger_to_promote` threshold has been reached
   - User-provided input (bug report, feature request, incident summary)

2. Document the observation (§21.6):
```yaml
architecture_amendment:
  reason: "Phase 6b iterate — measured signal"
  observation: "<one-line description of the signal>"
  signal_value: "<metric: latency p95 = 820ms for 15min>"
  pattern_to_promote: "§20.X" # if deferred pattern applies
```

3. Propose the change — emit the amendment YAML diff (not a full re-plan)

4. **🔒 Human Checkpoint** — present amendment, wait for approval

5. On approval:
   - Run the implementation through the Phase 3 self-review protocol (build → test → validate → smoke → audit)
   - Update `patterns_applied` / `patterns_deferred`
   - Re-run Phase 4 critical flows
   - Commit at each passing milestone

Append to `operate-log.yaml`:
```yaml
- cycle: iterate
  phase: "6b"
  date: <ISO-8601>
  trigger: "<signal description>"
  amendment: "<pattern or change>"
  status: proposed | approved | rejected | implemented
  action_required: true
  tokens_used: <count>
```

---

## Phase 6c — Maintain

Run weekly. Human is only notified if `action_required: true`.

Run each step in order. Stop at first ABORT condition (same rules as Phase 3 guardrails: tool exit code outside [0,1], scan tool parse error).

**Step 1 — Dependency scan:**
```bash
# Go
govulncheck ./...
# Bun/TS
npm audit --audit-level=high
```
- Clean: log and continue
- Findings: classify severity; HIGH → escalate immediately; MEDIUM → propose patch PR; LOW → log only

**Step 2 — Security scan:**
```bash
# Go
gosec -fmt=json -out=sec-report.json ./...
gitleaks detect --report-path=secrets-report.json
# Bun/TS
npx semgrep --config=auto --json > sec-report.json
gitleaks detect --report-path=secrets-report.json
```
- HIGH/CRITICAL → escalate; MEDIUM → include in report; LOW → log

**Step 3 — Secret rotation check (§25.3):**
Read `security-log.yaml` for last rotation dates. Flag any secret past its rotation schedule.

**Step 4 — Backup/restore drill (monthly only, skip if last drill was <28 days ago):**
- Restore latest Postgres snapshot to staging
- Verify row counts within 5% of production
- Run `make smoke-test` against restored staging
- Log result

**Output — post to chat:**
```yaml
- cycle: maintain
  phase: "6c"
  date: <ISO-8601>
  dep_scan: clean | "<N findings>"
  security_scan: clean | "<N findings>"
  secret_rotation: clean | "due: <key-name> in <N> days" | "rotated: <key-name>"
  backup_drill: pass | fail | skipped
  action_required: true | false
  tokens_used: <count>
```

If `action_required: false` — post the YAML block and stop (no recap paragraphs). Human does not need to respond.
If `action_required: true` — post the YAML block, describe the specific action needed, wait for human instruction.
