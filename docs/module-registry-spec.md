# Module Registry Specification

> For contributors adding new modules or implementations to the registry.

---

## What is the Module Registry?

The module registry is a shared library of pre-built, swappable components for SaaS web applications. It is the single source of truth for:

- **Module contracts** — Go and TypeScript interfaces defining what each capability looks like
- **Module implementations** — concrete implementations of those contracts (Clerk, Stripe, Resend, etc.)
- **Module manifests** — metadata files (`module.yaml`) telling agents how to use each module
- **Project templates** — starter structures for Go and Bun backends
- **Orchestration documents** — the playbook and rulebook that agents follow

When a new project is started, the agent uses `modkit init` to pull from this registry and assemble the project.

---

## Repository Structure

```
module-registry/
├── orchestration/
│   ├── playbook.md              ← agent workflow (6 phases)
│   ├── composition-rulebook.md  ← wiring rules
│   └── registry.yaml            ← master module index
├── contracts/
│   ├── go/                      ← Go interfaces
│   └── ts/                      ← TypeScript interfaces
├── modules/
│   ├── {module-name}/
│   │   ├── module.yaml          ← module manifest
│   │   ├── config.schema.json   ← env var schema
│   │   ├── docs/
│   │   │   ├── AGENT.md         ← agent-facing docs
│   │   │   └── README.md        ← human-facing docs
│   │   ├── impl/
│   │   │   ├── {provider}-go/   ← Go implementation
│   │   │   └── {provider}-ts/   ← TypeScript implementation
│   │   └── tests/
│   │       └── contract_test.go ← interface compliance tests
├── templates/
│   ├── project-go/              ← Go project template
│   └── project-bun/             ← Bun project template
├── modkit/                      ← CLI tool (Go)
└── docs/
    ├── module-registry-spec.md  ← this file
    └── modkit-cli-spec.md       ← CLI specification
```

---

## Module Structure

Every module must contain these files:

### `module.yaml` — Module Manifest

```yaml
name: "email"
version: "1.0.0"
description: "Transactional email delivery"
category: "notification"   # auth | payments | storage | notification | cache | realtime | search | observability | feature-flags | jobs | error-tracking | cicd
phase: "mvp"               # mvp | v2 | later

# What this module provides
interface_go: "contracts/go/email.go"
interface_ts: "contracts/ts/email.ts"

# What this module needs
dependencies:
  required: []
  optional:
    - "observability"
    - "jobs"

# Always include regardless of selection
always_include: false

# Configuration schema
config_schema: "config.schema.json"

# Default implementation
default_impl: "resend"

# Available implementations
implementations:
  - name: "resend"
    label: "Resend"
    phase: mvp
    runtimes:
      go: "impl/resend-go"
      bun: "impl/resend-ts"
  - name: "sendgrid"
    label: "SendGrid"
    phase: v2
    runtimes:
      go: "impl/sendgrid-go"
      bun: "impl/sendgrid-ts"

# Agent context file
agent_docs: "docs/AGENT.md"
```

### `config.schema.json` — Configuration Schema

JSON Schema defining all environment variables this module requires:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "EMAIL_PROVIDER": {
      "type": "string",
      "enum": ["resend", "sendgrid"],
      "description": "Email service provider"
    },
    "EMAIL_API_KEY": {
      "type": "string",
      "description": "API key for the email provider",
      "sensitive": true
    },
    "EMAIL_FROM_DEFAULT": {
      "type": "string",
      "format": "email",
      "description": "Default from address"
    }
  },
  "required": ["EMAIL_PROVIDER", "EMAIL_API_KEY", "EMAIL_FROM_DEFAULT"]
}
```

Mark secrets with `"sensitive": true` — `modkit init` uses this to generate `.env.example` with placeholder values.

### `docs/AGENT.md` — Agent Instructions

Written specifically for agents. Uses clear, imperative language:

```markdown
# Email Module — Agent Instructions

## When to use
Include this module when the project requires transactional emails:
- User signup confirmation
- Password reset / magic links
- Invoice or receipt delivery
- Status change notifications

Do NOT use for marketing/bulk emails (different concern, different tooling).

## How to wire

### Go
1. Import `contracts.EmailService` interface
2. Initialize in bootstrap: `email := resend.New(cfg.Email)`
3. Inject into handlers that need email capability
4. For async sending, always pair with the jobs module

### Bun
1. Import `IEmailService` from `contracts/ts/email.ts`
2. Initialize: `const email = new ResendEmailService(config.email)`
3. Inject into handlers

## Common patterns

### Send an email
```go
result, err := email.Send(ctx, contracts.EmailMessage{
    To:      []string{"user@example.com"},
    From:    "noreply@yourapp.com",
    Subject: "Your invoice is ready",
    Body:    contracts.EmailBody{HTML: "<p>...</p>", Text: "..."},
})
```

### Async email (preferred for non-critical emails)
Pair with the jobs module to send emails asynchronously:
```go
jobs.Enqueue(ctx, "user:send_welcome_email", UserID{ID: user.ID})
// In job handler: email.Send(...)
```

## Required env vars
See `config.schema.json` for the full list.

## Do NOT
- Store API keys in code — use config only
- Send marketing/bulk email through this module
- Ignore email delivery failures — always log errors
```

### `docs/README.md` — Human Documentation

Standard markdown for humans: overview, setup instructions, provider comparison, examples.

### `impl/{provider}-{runtime}/` — Implementation Stubs

Each implementation directory must contain:

**Go implementation (`impl/resend-go/`):**
```
resend-go/
├── resend.go           ← implementation (satisfies contracts.EmailService)
├── resend_test.go      ← unit tests
└── README.md           ← how to configure this implementation
```

**TypeScript implementation (`impl/resend-ts/`):**
```
resend-ts/
├── resend.ts           ← implementation (satisfies IEmailService)
├── resend.test.ts      ← unit tests
└── README.md
```

### `tests/` — Contract Tests

```
tests/
├── contract_test.go    ← Go: runs all implementations against the interface
└── contract.test.ts    ← TS: same for TypeScript
```

---

## Adding a New Module

1. **Define the interface** in `contracts/go/{module}.go` and `contracts/ts/{module}.ts`
2. **Create the module directory** `modules/{module-name}/`
3. **Write the manifest** `module.yaml` with phase, category, dependencies
4. **Write the config schema** `config.schema.json`
5. **Write agent docs** `docs/AGENT.md` — be specific about when to use and how to wire
6. **Create implementation stubs** in `impl/{provider}-go/` and `impl/{provider}-ts/`
7. **Write contract tests** in `tests/`
8. **Add to `registry.yaml`** under `modules:`
9. **Update `composition-rulebook.md`** if the module has wiring requirements

---

## Adding a New Implementation

1. Create `impl/{provider}-{runtime}/` in the relevant module directory
2. Implement the module's contract interface
3. Write tests
4. Add to `module.yaml` under `implementations:`
5. Add to `registry.yaml` under the module's `implementations:`

---

## Module Phases

| Phase | Meaning |
|-------|---------|
| `mvp` | Include in MVP scaffold by default selection rules |
| `v2` | Available and stable but not in MVP by default |
| `later` | Planned but not yet implemented |

Do not add `later`-phase modules to `registry.yaml` until they are implemented.

---

## Versioning

Module versions follow semver (`{major}.{minor}.{patch}`):
- `patch` — bug fixes, no interface changes
- `minor` — new optional interface methods (backward compatible)
- `major` — breaking interface changes (rare — requires updating all implementations)

When a module interface changes, update all implementations in the same PR.
