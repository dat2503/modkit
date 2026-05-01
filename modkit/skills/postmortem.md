---
description: Guide a blameless postmortem from an incident and propose follow-up architecture amendments
---

You are running a guided postmortem for a production incident. Load context once (§18):
- `~/.modkit/cache/orchestration/composition-rulebook.md` §23.4 (postmortem template)
- `~/.modkit/cache/orchestration/composition-rulebook.md` §21.6 (evolution protocol)
- The project's `slo.yaml` (SLO targets for severity classification)
- The project's `patterns_deferred` block from `architecture-plan.yaml`

---

## Step 1 — Gather incident inputs

Ask the user for (or read from context if already provided):
- Incident start and end time
- Affected service/endpoints
- How the incident was detected (alert, user report, internal discovery)
- Link to Sentry issue(s) or error logs if available

If log links are provided, read them (cheapest source first: structured log excerpt > Sentry event JSON > screenshot).

---

## Step 2 — Classify severity

Using `slo.yaml` targets:
- **P1**: SLO breach, data loss, security event, or payment disruption
- **P2**: Degraded performance (latency above SLO target, elevated error rate), no data loss
- **P3**: Non-user-facing issue, caught before widespread impact

---

## Step 3 — Build timeline

Construct a timeline from the gathered inputs. Ask the user to fill in gaps. Keep entries terse — time + one-line event.

---

## Step 4 — Root cause analysis

Ask: "What was the proximate cause? What was the systemic cause that allowed the proximate cause to have impact?"

Proximate: the specific thing that failed.
Systemic: the gap in process, monitoring, testing, or architecture that let it reach production or users.

Frame this as "the system allowed X" not "person Y did X."

---

## Step 5 — Emit postmortem YAML (§23.4 template)

```yaml
incident_id: "INC-<YYYYMMDD>-<N>"
date: "<ISO-8601>"
severity: P1 | P2 | P3
duration_minutes: <N>
summary: "<one sentence: what failed, what impact>"
timeline:
  - time: "<HH:MM>"
    event: "<what happened>"
root_cause: "<systemic cause>"
contributing_factors: []
what_went_well: []
action_items:
  - description: ""
    owner: "agent | human"
    due_date: ""
patterns_to_promote: []
```

---

## Step 6 — Check for architecture amendments

Compare the root cause and contributing factors against:
1. `patterns_deferred` — did a deferred pattern's trigger fire? If yes, list it in `patterns_to_promote`
2. §20 DDIA patterns — is there an applicable pattern not yet in `patterns_applied`?
3. §23.2 alert rules — was detection delayed because an alert was missing?
4. §23.3 runbooks — did the lack of a runbook slow mitigation?

For each gap found, propose an architecture amendment (§21.6 Architectural Feedback Loop format).

---

## Step 7 — Present and wait

Post the completed postmortem YAML to chat. List action items clearly with owners.

If any `patterns_to_promote` are identified — emit the corresponding architecture amendment YAML and wait for human approval before implementing (§21.6 evolution protocol).

Save the postmortem YAML to `postmortems/<incident_id>.yaml` in the repo.

---

## Token rule

The postmortem is a structured artifact. Keep each field terse. No narrative prose outside YAML. Total output target: ≤200 tokens for the YAML block + ≤100 tokens for the amendment proposals.
