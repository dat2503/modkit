---
description: Start a new web application using the modkit module registry
---

You are starting a new web application project. Follow the modkit 6-phase playbook.

## Step 1 — Load context

Load these files **ONCE per session** — do not re-read them in later steps (§18 token efficiency):
1. `~/.modkit/cache/docs/agent-instructions.md`
2. `~/.modkit/cache/orchestration/playbook.md`
3. `~/.modkit/cache/orchestration/composition-rulebook.md`
4. `~/.modkit/cache/orchestration/registry.yaml`

Load module `AGENT.md` only for selected modules (after Phase 1 approval), not the full catalog.

## Step 2 — Intake (Phase 0)

Ask the user about their idea. Produce a structured brief in YAML format:

```yaml
project_name: ""
domain: ""
runtime: go  # or bun
entities:
  - name: ""
    fields: []
user_roles:
  - role_name:
      - action 1
      - action 2
key_flows:
  - "User does X → Y happens → Z result"
assumptions:
  - "assumption 1"
```

Present the brief and **wait for approval** before continuing.

## Step 3 — Module selection (Phase 1)

Based on the brief, select modules from the registry:
- Apply the §21.3 lean defaults — observability and error-tracking are optional (skip for prototypes)
- For each candidate module, read its AGENT.md: `~/.modkit/cache/modules/{name}/docs/AGENT.md`
- Present a table:

| Module | Include? | Rationale |
|--------|----------|-----------|
| observability | production app? yes / prototype? skip | Adds tracing overhead |
| ... | | |

After finalizing selection, emit `toolchain.yaml` (§22.5):
```yaml
toolchain:
  modules: [auth/clerk, payments/stripe, ...]
  mcp_tools_available: [playwright, canva, github]
  runtime_libraries: [go-chi, sqlc, ...]
  external_services: [clerk, stripe, ...]
```

Present the module selection table **and** `toolchain.yaml`. **Wait for approval** before continuing.

## Step 4 — Scaffold (Phase 3)

Run modkit init with the approved selections:

```bash
modkit init \
  --name <project_name> \
  --runtime <runtime> \
  --go-module <go-module-path> \
  --modules <comma-separated-list> \
  --no-prompt \
  --registry ~/.modkit/cache
```

Then `cd` into the generated project.

## Step 5 — Architecture (Phase 2)

Design the application architecture:

1. **Database schema** — write SQL migrations for each entity
2. **API routes** — table with method, path, auth requirement, description
3. **Frontend pages** — table with route and description
4. **Module wiring** — bootstrap init order

Then apply the §21 + §22 layer:

5. **Capture project signals** (§22.3): scale estimate, team size, longevity, user type, regulatory context
6. **Pattern Selection** (§20 + §21 tier filter):
   - Tier-0: apply automatically (§20.1 payment/webhook, §20.10, §20.13, §20.17, §20.18, §20.20)
   - Tier-1: apply only if a concrete trigger exists in the brief — run the Reverse-YAGNI test
   - Tier-2: list in `patterns_deferred` with the measured trigger to promote them
   - Pattern budget: ≤8 from §20 excluding Tier-0
7. **Emit** `patterns_applied`, `patterns_deferred`, `mvp_profile_overrides`, `deviations`

Apply §21.5 forbidden list and §21.7 boring-stack principle. Reference composition-rulebook.md §19–§22 by ID — do not paraphrase.

Present the full architecture plan (schema + routes + pages + wiring + patterns) and **wait for approval**.

## Step 5.5 — Design Analysis (Phase 2.5)

Run this step if the user has provided **any** of: Canva URL, Excalidraw checkpoint, Figma URL, screenshot/image, design brief, reference URL, component library preference, style guide, accessibility requirements, or responsive specs.

1. Load assets using the cheapest source first (text brief > Excalidraw > Canva > image)
2. Produce `ui-spec.yaml` (brand, layout, pages, components, interactions, accessibility, responsive, component_library)
3. Map each frontend page from the architecture plan to design components; flag gaps
4. Present `ui-spec.yaml` and **wait for confirmation**

`ui-spec.yaml` is the ONLY design artifact carried into Step 6. Do not re-load original design files later.

**Skip this step** if no design assets were provided.

---

## Step 6 — Implement (Phase 3 continued)

Build the app logic on top of the scaffold in this order:

1. Infrastructure — `.env`, `make setup`, verify health endpoint → **Milestone 3a**
2. Database — migrations, run `make migrate` → **Milestone 3b**
3. Service layer + API handlers — wire into router with auth middleware → **Milestone 3c**
4. Frontend pages — components matching `ui-spec.yaml` (if Phase 2.5 ran) → **Milestone 3d**

**After each milestone**, run the Agent Self-Review Protocol (see Playbook):
1. `make build` — fix errors before moving on
2. `make test` — fix failures before moving on
3. `modkit validate --output json` — fix violations immediately
4. Run the milestone smoke tests (Playbook §Smoke Tests per Milestone)
5. Self-audit: re-read own diff against architecture plan and `ui-spec.yaml`

**Self-improvement loop:** fix → rerun up to 3 same-error / 6 different-error iterations, then escalate.

**On all checks passing:**
```bash
git add -A
git commit -m "milestone:{3a|3b|3c|3d} — {one-line description}"
git tag milestone/{name}/$(date +%Y%m%d-%H%M)
```
Post the self-review.log YAML block to chat. Continue to next milestone.

**Cross-milestone regression:** at the start of each milestone, re-run the previous milestone's smoke tests.

**Before writing custom code** for any capability, run the §22.1 reuse check. Emit `reuse_check` block. Watch for §22.6 "don't reinvent" smells.

**Human checkpoints:**
- Mandatory: Phase 0 brief, Phase 1 modules, Phase 2 architecture, Phase 2.5 design (if run), Phase 4 validation, Phase 5 deploy
- Optional: end of Milestone 3c if self-improvement loop triggered, architecture amended, or API surface >10 routes
- Agent-only (no human): every other sub-milestone when all checks pass cleanly

**Guardrails:** respect hard limits (§17) even if user says "keep going." Never self-resume after a guardrail trigger.

Follow composition-rulebook.md §19–§22 throughout.

## Step 7 — Operate (Phase 6)

After Phase 5 deploy succeeds, transition to ongoing operations via the `/operate` skill.

Phase 6 runs in three sub-phases:
- **6a** (one-time): emit `slo.yaml` and `alerts.yaml` — measurable targets before the system has time to regress
- **6b** (continuous): evolution loop — when a measured signal fires, propose an architecture amendment, get approval, implement
- **6c** (weekly): dependency scan + security scan + secret rotation check + monthly backup drill

**Reference:** Playbook Phase 6, composition-rulebook.md §23–§26.

**This step has no end** — `/operate` can be re-run at any time to check 6c status or respond to a 6b trigger.
