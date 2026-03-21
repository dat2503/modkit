---
description: Start a new web application using the modkit module registry
---

You are starting a new web application project. Follow the modkit 6-phase playbook.

## Step 1 — Load context

Read these files in order:
1. ~/.modkit/cache/docs/agent-instructions.md
2. ~/.modkit/cache/orchestration/playbook.md
3. ~/.modkit/cache/orchestration/registry.yaml

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
- Always include: observability, error-tracking
- For each candidate module, read its AGENT.md: `~/.modkit/cache/modules/{name}/docs/AGENT.md`
- Present a table:

| Module | Include? | Rationale |
|--------|----------|-----------|
| observability | always | Required |
| ... | | |

**Wait for approval** before continuing.

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

Present the architecture plan and **wait for approval**.

## Step 6 — Implement (Phase 3 continued)

Build the app logic on top of the scaffold in this order:

1. Infrastructure — `.env`, `make setup`, verify health endpoint
2. Database — migrations, run `make migrate`
3. Service layer — business logic structs and methods
4. API handlers — wire into router with auth middleware
5. Frontend pages — Next.js pages with API client

After each group:
- Run `go build ./...` (or `bun build`)
- Run `/validate` to check wiring
- **Stop for user review**

Follow the composition-rulebook.md rules throughout.
