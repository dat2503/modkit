# Workflow Improvements — Agent Self-Review, Design Analysis, Guardrails & Engineering Principles

This document explains the additions made to the modkit agent workflow. Read this before reading the updated `playbook.md` and `composition-rulebook.md`.

---

## Why These Changes Were Made

The original 6-phase playbook gave agents a workflow, but left gaps:

- Phase 3 had no structured self-review — agents ran one build command and kept going
- No self-improvement loop with bounded retries or escalation
- No design analysis phase — frontends were built from text descriptions only
- No guardrails — nothing stopped runaway agents or spiraling fix loops
- No DDIA/engineering principles — patterns were applied inconsistently or over-eagerly
- Token usage was unconstrained — agents re-read files, produced verbose logs

These additions close those gaps while preserving the core 6-phase structure.

---

## Overview of Changes

### Playbook Changes

| Addition | Purpose |
|----------|---------|
| Agent Self-Review Protocol | Mandatory 5-step verification before each milestone |
| Agent Guardrails & Loop Control | Hard limits on retries, file writes, wall time, tokens |
| Phase 2.5 — Design Analysis | Extract ui-spec.yaml from Canva, Excalidraw, Figma, or design briefs before coding |
| Phase 2 Pattern Selection | Agents declare which §19/§20 patterns apply and why — reviewed at human checkpoint |
| Phase 2 Project-Fit signals | Classify project scale/team before choosing patterns |
| Phase 1 Toolchain step | Emit toolchain.yaml after module selection — drives reuse in Phase 3 |
| Architectural Feedback Loop | When architecture is wrong mid-Phase 3, amend with approval — never silently hack around |
| Git Checkpoints | Commit at every passing milestone; never commit failed work |
| Optional Phase 3c Checkpoint | Human can review API before frontend work in complex cases |
| Strengthened Phase 4 | Real E2E: docker-compose up + critical-flow walk via curl + Playwright |
| Self-Review Telemetry | ~50-token YAML log per milestone replaces verbose chat narration |
| Phase 2 Lean Filter (§21) | Apply Tier-0/1/2 pattern tiers + Reverse-YAGNI test at architecture approval |

### Rulebook Additions

| Section | Content |
|---------|---------|
| §14 Agent Verification Loop | Per-handler build+validate + self-audit checklist |
| §17 Guardrails & Loop Control | Summary of hard limits and escalation protocol |
| §18 Token Efficiency | Load-once rules, output format rules, cheap-first verification |
| §19 Engineering Principles | SOLID, DRY, YAGNI, Separation of Concerns, 12-Factor, coding principles |
| §20 Data-Intensive Patterns | 21 DDIA-derived patterns each with trigger, pattern, DDIA chapter ref |
| §21 Lean MVP & Anti-Over-Engineering | Pattern tiers, Reverse-YAGNI test, lean defaults, evolution protocol |
| §22 Reuse Over Reinvent + Project Fit | Reuse hierarchy, library rubric, deviation docs, "don't reinvent" smell list |

---

## Core Principles (read this before the details)

### Token Efficiency First
Every rule is designed to minimize token usage. Cheap verification before expensive. Structured YAML not prose. Load context once. Reference by §ID not paraphrase. Self-review logs are ~50 tokens not paragraphs.

### Patterns Are a Menu, Not a Requirement
§20 has 21 patterns. A typical MVP applies 6–8. The §21 tier system + Reverse-YAGNI test prevents over-application. Patterns are pulled by evidence (real triggers in the brief), never pushed by speculation.

### Project Fit Over Best Practice
"Best practice" assumes a generic project. The right choice for a 10-user tool differs from a multi-tenant SaaS. §22 requires agents to capture project signals (scale, team, longevity, regulation) before applying any pattern. Deliberate deviations are documented — first-class, not exceptions.

### Reuse Before Building
§22 defines a mandatory reuse hierarchy: modkit modules → MCP tools → stdlib → mature library → custom. Custom code requires a reuse_check block explaining what was checked and why it didn't fit.

---

## Agent Self-Review Protocol

Every milestone (3a, 3b, 3c, 3d, Phase 4) must pass this 5-step sequence before being declared complete:

1. **Build** — exit code only; parse output only on failure
2. **Test** — pass/fail count; full output only on failure
3. **Wiring** — `modkit validate --output json`; parse JSON
4. **Smoke** — milestone-specific HTTP checks via curl (status codes only)
5. **Self-audit** — re-read own diff against the approved architecture plan

**Self-improvement loop:**
- Same-error retries: max 3 (identical first 100 chars of error message)
- Different-error retries: max 6 (each iteration diagnoses a genuinely new cause)
- Same-fix-same-error: STOP immediately (no count)
- Circular A→B→A dependency: STOP after detecting once

After exhausting limits → 200-token failure report to human → wait for guidance.

---

## Guardrails Reference

| Limit | Value | Action |
|-------|-------|--------|
| Same-error retries per milestone | 3 | Escalate |
| Different-error retries per milestone | 6 | Escalate |
| File writes per Phase 3 sub-phase | 50 | Stop, present |
| Consecutive tool failures | 5 | Stop |
| Phase 3 wall time | 90 min | Stop, summarize |
| Token budget exceeded | See §Guardrails | Warn 80%, stop 100% |

**Abort conditions (no retry):** partial migration failure, build tool exit code outside [0,1], DROP/DELETE not in approved plan, `modkit validate` parse error.

---

## Phase 2.5 — Design Analysis

Triggers when user provides: Canva URL, Excalidraw checkpoint, Figma URL, screenshot/image, design brief, reference URL, component library preference, style guide, accessibility requirements, or responsive specifics.

Agent loads assets (cheapest source first) and produces `ui-spec.yaml`:
```yaml
design_source: "canva://DAB123"
brand:
  primary_color: "#1A73E8"
  font_family: "Inter, sans-serif"
layout:
  nav: "top fixed"
  sidebar: "left, 240px"
pages:
  - route: /dashboard
    components: [StatCard, InvoiceTable]
components:
  - name: StatCard
    style: "white card, rounded-lg, subtle shadow"
```

`ui-spec.yaml` is the ONLY design artifact carried into Phase 3d. Original design files are not re-read.

---

## §19 Engineering Principles (Summary)

| Principle | Key rule |
|-----------|---------|
| §19.1 SOLID | One handler = one thing; contracts are small; depend on interfaces |
| §19.2 DRY | Extract after 3rd occurrence |
| §19.3 YAGNI | No field, route, or module not in approved plan |
| §19.4 Separation of Concerns | handler → service → repository → DB; no shortcuts |
| §19.5 Twelve-Factor | Config via env; stateless; stdout logs; graceful shutdown; dev/prod parity |
| §19.6 Coding Principles | Fail fast; errors as values; pure functions; composition; <50-line functions |

---

## §20 Data-Intensive Patterns (Summary)

Patterns are grouped by problem and each has a concrete trigger — apply only when the trigger is true.

**Reliability (Tier 0/1):** idempotency keys, outbox, at-least-once jobs, retry+backoff, circuit breaker
**Scalability (Tier 1/2):** cache-aside (§7), read replicas, cursor pagination, queue backpressure
**Maintainability (Tier 0/1):** backward-compat migrations, versioned APIs, structured logs
**Data Integrity (Tier 0/1):** transactions/isolation, optimistic concurrency, event ordering, soft deletes (§9)
**Distributed Hygiene (Tier 0):** webhook sig verification, timeouts everywhere, rate limiting
**Performance (Tier 0/1):** N+1 prevention, index by query pattern

Full pattern specs in `composition-rulebook.md §20`.

---

## §21 Lean MVP — Pattern Tiers

| Tier | Rule | Patterns |
|------|------|---------|
| Tier 0 — Safety | Always apply | §20.1 (payment/webhook only), §20.10, §20.13, §20.17, §20.18, §20.20 |
| Tier 1 — Direct trigger | Apply only when brief contains a concrete trigger | §20.2, §20.3, §20.4, §20.8, §20.11, §20.21 |
| Tier 2 — Deferred | Never at MVP; add when measured signal fires | §20.5, §20.7, §20.9, §20.14, §20.15, §20.19 (most routes) |

**Reverse-YAGNI test:** For every Tier-1 pattern, answer "what breaks today without this?" If "nothing today" → defer.

**Pattern budget:** ≤8 from §20 (excluding Tier-0) per MVP.

**Lean defaults (§21.3):** monolith, single Postgres, polling not WebSocket, ILIKE not Elasticsearch, one IdP, no CDN, no multi-region.

**Forbidden at MVP (§21.5):** microservices, message buses, event sourcing, CQRS, service mesh, gRPC internal, multi-DB, premature sharding.

**Evolution protocol (§21.6):** observe measured signal → document → propose amendment → human approves → implement. Patterns pulled by evidence, never pushed by speculation.

---

## §22 Reuse Hierarchy

Check in order before writing custom code:
1. Existing modkit module
2. MCP tools available in session
3. Runtime standard library
4. Well-maintained third-party library (maintained, licensed, semver stable, bounded deps)
5. Custom implementation (emit reuse_check block explaining why 1–4 didn't fit)

**Don't-reinvent list:** HTTP retry, HMAC verification, JWT parsing, UUID generation, date/time math, rate limiting, CSV/file MIME/HTML sanitization, markdown rendering, image resizing, email templates, crypto primitives.

---

## Git Checkpoint Protocol

After each passing milestone:
```bash
git add -A
git commit -m "milestone:{3a|3b|3c|3d|phase4} — {one-line description}"
git tag milestone/{name}/$(date +%Y%m%d-%H%M)
```

On milestone failure: do NOT commit. Show `git status` + `git diff --stat` to human. Human decides: discard, keep dirty, or accept partial.

---

## Phase 4 — End-to-End Validation

Full E2E sequence (cheap → expensive):
1. Static: `make build`, `make test`, `modkit validate`
2. Services: `docker-compose up -d` + poll healthchecks + `make migrate`
3. Critical flows: curl (API) + Playwright MCP (frontend) for each `key_flow` from project-brief.yaml
4. Design check (if Phase 2.5 ran): extract CSS variables, compare to ui-spec.yaml brand

Output: compact `validation-report.yaml` with pass/fail per flow. Full logs to file, not chat.

---

## Phase 6 — Operate (Day-2 SDLC)

### Why Phase 6 Was Added

The original playbook covered the **build half** of SDLC well — requirements, design, implementation, testing, deploy. What it left out was the **operate half**: what happens after the first ship. Traditional SDLC covers this with ceremonies (sprints, retros, standups). Modkit replaces ceremony with **agent-runnable Day-2 protocols** that produce structured artifacts and only page a human when there's a real finding.

**What we keep from traditional SDLC:** the substance — incident response, release management, security hygiene, data governance.

**What we drop:** the ceremony — sprints, retros, standups, RACI matrices, status reports. These don't add value when an agent does the work.

### Phase 6 Sub-phases

| Sub-phase | When | What |
|-----------|------|------|
| 6a Instrument | Once after first deploy | Emit `slo.yaml`, `alerts.yaml`, `data-governance.yaml`, `threat-model.yaml`, `pii-inventory` |
| 6b Iterate | On measured trigger | Evolution loop: observe signal → propose amendment → approve → implement |
| 6c Maintain | Weekly | Dep scan, security scan, secret rotation check, monthly backup drill |

### New Rulebook Sections

| Section | Content |
|---------|---------|
| §23 Day-2 Operations | SLI/SLO definitions, alert design rules, runbook template, postmortem template (§23.4), lean oncall |
| §24 Release Management | Conventional Commits, semver rules, rollback protocol (§24.3), canary via feature flags, release checklist |
| §25 Security Lifecycle | CI scan suite (govulncheck/gosec/semgrep/gitleaks), secret rotation cadence, threat model template, quarterly review checklist |
| §26 Data Governance | Retention table, GDPR export/delete endpoints, PII inventory, backup/restore drill |

### New Skills

| Skill | Purpose |
|-------|---------|
| `/operate` | Phase 6 entry point — detects which sub-phase to run, walks through 6a/6b/6c |
| `/postmortem` | Guided blameless postmortem builder — severity classification, timeline, root cause, §21.6 amendment proposals |
| `/release` | Pre-release checklist → semver bump → CHANGELOG (via git-cliff/changesets) → tag → deploy → verify |

### Token Efficiency

Phase 6 follows the same §18 token rules as the rest of the playbook:
- operate-log.yaml entries are ~50 tokens (fixed schema, no narrative)
- Phase 6c only pages a human if `action_required: true` — clean runs are silent
- Phase 6a emits YAML artifacts, not prose explanations
- `/postmortem` outputs ≤200 tokens for the YAML block + ≤100 for amendment proposals

---

## Cross-Project Learning Loop (§27)

### Why Added

Per-project artifacts (self-review.log, postmortems, patterns_deferred) capture lessons but they never leave the project. Every new project started from zero. Mistakes from project A didn't inform agents on project B.

### What Was Added

**`learnings/catalog.yaml`** — a registry-level cross-project lessons store. Each entry: id, category (bug/anti-pattern/integration-gotcha/etc.), domains, title, lesson text, trigger condition, source, date.

**`/learn` skill** — run after a postmortem or at end of project. Agent extracts 0–3 generalizable lessons, applies §27.2 qualification filter (would it change a Phase 2 decision? is it non-obvious? is it generalizable?), checks for catalog duplicates, proposes entries. Human approval required before any write.

**`/new-app` updates** — catalog loaded in Step 1 (load-once, §18). At Step 5 architecture, agent scans catalog for ≤5 relevant entries and surfaces them as "Heads up from past projects" before pattern selection.

**§27 in composition-rulebook.md** — entry schema, qualification rules, query protocol, hygiene (18-month prune, 200-entry cap, no-rulebook-duplication rule).

---

## CI/CD Compliance Postures (§28)

### Why Added

Forcing enterprise CI/CD practices onto every project is over-engineering — a solo builder doesn't need SBOM + cosign + CODEOWNERS. But those practices matter when the project grows. §28 scales CI/CD with the project, using the same evidence-driven logic as §21's pattern tiers.

### Three Postures

| Posture | Auto-selected when | Additions beyond §24/§25 baseline |
|---------|-------------------|----------------------------------|
| `solo` | team=1 AND <100 users | None — baseline SAST/SCA/secrets is sufficient |
| `startup` | small team OR public users | Dependabot, PR template, Trivy scan, Codecov 60%, branch protection script |
| `enterprise` | regulated OR B2B SaaS OR enterprise customers | Everything in startup + SBOM (syft), container signing (cosign), license compliance, CODEOWNERS, issue templates, Codecov 80% |

Posture is auto-suggested from §22.3 project signals at Phase 2 and confirmed at the Phase 1 checkpoint. User can override. Promotes via §21.6 evolution protocol — trigger observed, amendment proposed, human approves.

### What Was Added

- **`modules/cicd/templates/startup/`** — dependabot.yml, PULL_REQUEST_TEMPLATE.md, codecov.yml (60%), setup-branch-protection.sh
- **`modules/cicd/templates/enterprise/`** — CODEOWNERS, bug_report.md, feature_request.md, codecov.yml (80%)
- **`modules/cicd/docs/AGENT.md`** — Compliance Posture section documenting what CI job additions each posture generates (Trivy/codecov for startup; SBOM/cosign/license for enterprise)
- **Contracts updated** — `CompliancePosture` field in `CICDConfig` (Go + TS)
- **Config schema updated** — `COMPLIANCE_POSTURE` enum added

---

## User-Driven Module Selection

### Why Added

Previously the agent picked modules and presented a table for approval. The user wanted explicit agency — "tick what you want."

### What Changed

**`/new-app` Step 3** now presents the full module menu (all available modules, with descriptions and trigger conditions) and explicitly asks the user for yes/no per module. Compliance posture is selected alongside. Agent honors every explicit choice; auto-includes required dependencies and explains why.

**`/configure` skill** — new. For post-scaffold module changes on an existing project. Shows current state, validates dependency chains, performs impact analysis for removals (scanning usages in code), presents change plan with human approval gate, implements with self-review protocol per module, and re-evaluates posture on completion.
