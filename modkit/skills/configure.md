---
description: Add or remove modules from an existing modkit project post-scaffold
---

You are modifying the module configuration of an already-scaffolded modkit project. Load context once (§18):
- `~/.modkit/cache/orchestration/composition-rulebook.md` (wiring rules §3–§7)
- `~/.modkit/cache/orchestration/registry.yaml` (available modules)
- Current project's `toolchain.yaml` (what's already included)
- Current project's `architecture-plan.yaml` (approved patterns and API surface)

Do NOT re-read files already loaded in this session.

---

## Step 1 — Show current configuration

Read `toolchain.yaml` and present the current module state:

```
Currently included:
  ✓ auth/clerk
  ✓ cache/redis
  ✓ cicd/github-actions

Available to add:
  ○ payments/stripe      — process payments and subscriptions
  ○ email/resend         — transactional email
  ○ storage/s3           — file uploads and asset storage
  ○ jobs/asynq           — background job processing (requires cache)
  ○ observability        — OpenTelemetry tracing and structured logging
  ○ error-tracking       — Sentry error capture and alerting
  ○ search               — full-text search (Postgres trigram at MVP)
  ○ feature-flags        — feature gating and phased rollouts
  ○ realtime             — WebSocket connections (polling-first default)

Currently included (cannot remove without impact analysis):
  ✓ auth/clerk           — removing requires stripping all auth middleware
```

Ask: **"What would you like to change? You can say 'add payments', 'remove observability', or list multiple changes."**

Wait for the user's response before proceeding.

---

## Step 2 — Validate changes

For each requested change:

**Adding a module:**
1. Check `registry.yaml` for `dependencies.required` — if the module requires a dep not yet installed, include it automatically and tell the user
2. Check `dependencies.optional` — surface any optional dependencies the user might want alongside
3. Check §21.3 lean defaults — if the module is marked optional at prototype scale, confirm the user wants it
4. Check §21.5 forbidden list — if the request would introduce a forbidden pattern (e.g., adding a message bus instead of jobs/asynq), STOP and explain

**Removing a module:**
1. Scan the project's existing code for usages of the module's contract interface
2. Emit a removal impact list: "Removing X will affect: [list of files/routes/services]"
3. Confirm removal plan follows §20.10 (backward-compatible migration — add removal PR after verifying no usages)
4. **Never remove a module that has data-owning migrations already applied to production** without an explicit archival plan

---

## Step 3 — Present the change plan

Present a compact change plan before touching any code:

```yaml
configure_plan:
  add:
    - module: payments/stripe
      impl: stripe
      runtime: go
      requires: []
      wiring_steps:
        - "Add STRIPE_SECRET_KEY and STRIPE_WEBHOOK_SECRET to .env"
        - "Add PaymentsService to Deps struct in bootstrap.go"
        - "Wire after cache, before any payments routes"
        - "Add idempotency_keys migration (§20.1 — Tier-0 for payment routes)"
  remove: []
  compliance_posture_impact:
    - "Adding payments upgrades recommended posture to startup (adds Trivy + PR template)"
```

**🔒 Human checkpoint** — wait for approval before any code changes.

---

## Step 4 — Implement

On approval, implement in this order:

1. **Run `modkit init` for added modules** (or document manual wiring steps if modkit init can't add to existing project)
2. **Follow the §3–§7 wiring rules** for each added module
3. **Run the Agent Self-Review Protocol** after each module is wired:
   - `make build`
   - `make test`
   - `modkit validate --output json`
4. **Update `toolchain.yaml`** to reflect the new module list
5. **Update `architecture-plan.yaml`** — add any new patterns required (e.g., §20.1 for payments)
6. **Git commit** per milestone: `git commit -m "configure: add {module-name}"`

---

## Step 5 — Compliance posture check

After all module changes:
1. Re-evaluate `compliance_posture` based on updated project signals + new modules
2. If posture should change (e.g., adding payments → startups should at minimum add Trivy), emit an architecture amendment per §28.3
3. **🔒 Human checkpoint** on any posture upgrade — present the additions it entails

---

## Guardrails

- Never add a module marked `phase: later` in registry.yaml — those are not yet implemented
- Never add >2 modules in a single `/configure` run — complex multi-module wiring must be sequential
- If any `make build` fails after a wiring step, STOP and fix before the next module (never build on broken wiring)
- If the user requests a module that requires a new external service (Stripe, Resend, etc.) — remind them to create the account and set up keys before implementing
