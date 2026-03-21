# Module Registry — Claude Code Instructions

This is the **module registry** monorepo. It is a library that agents pull from to scaffold SaaS applications — not an application itself.

---

## Repository Layout

```
orchestration/   ← Read these FIRST before any work
  playbook.md           ← 6-phase agent workflow (authoritative)
  composition-rulebook.md ← All wiring rules (authoritative)
  registry.yaml         ← Master module index

contracts/
  go/            ← Go interfaces (one file per module)
  ts/            ← TypeScript interfaces (one file per module)

modules/
  {name}/
    module.yaml          ← Module manifest
    config.schema.json   ← Env var schema
    docs/AGENT.md        ← Agent-facing instructions
    docs/README.md       ← Human-facing docs
    impl/{provider}-go/  ← Go implementation
    impl/{provider}-ts/  ← TypeScript implementation
    tests/               ← Contract tests

templates/
  project-go/    ← Go project scaffold template
  project-bun/   ← Bun project scaffold template

modkit/          ← CLI tool (Go + Cobra)
docs/            ← Contributor guides
```

---

## Before Any Work

Load these documents in order:
1. `orchestration/playbook.md` — the 6-phase workflow
2. `orchestration/composition-rulebook.md` — all wiring rules
3. `orchestration/registry.yaml` — available modules

---

## Key Rules

### Contracts
- Go interfaces go in `contracts/go/{module}.go`
- TypeScript interfaces go in `contracts/ts/{module}.ts`
- All Go methods: `context.Context` as first parameter, return `(result, error)`
- All TS methods: `async`, return `Promise<T>`
- Never import concrete implementations in contracts

### Modules
- Every module needs: `module.yaml`, `config.schema.json`, `docs/AGENT.md`, `docs/README.md`
- `docs/AGENT.md` must include: when to use, how to wire (Go + Bun), common patterns, do-nots
- Implementations go in `impl/{provider}-{runtime}/` — named as `{provider}-go` or `{provider}-ts`
- Config is always injected via constructor — never read env vars directly in module code

### Phases
- `phase: mvp` — included in MVP by default selection rules
- `phase: v2` — available but not in MVP default
- `phase: later` — planned, not yet implemented (do not add to registry.yaml until implemented)

### Module Initialization Order
When wiring modules in any project:
1. observability (always first)
2. error-tracking (always second)
3. cache (before auth and jobs)
4. auth (after cache)
5. remaining modules in any order

### CLI (modkit)
- Source is in `modkit/` — Go + Cobra
- All commands support `--output json` and `--no-prompt` global flags
- Exit codes: 0=success, 1=general error, 2=validation failure, 3=missing dep, 4=config error, 5=network error, 6=runtime mismatch
- See `docs/modkit-cli-spec.md` for full spec

---

## Adding a Module (Checklist)

1. Define Go interface: `contracts/go/{module}.go`
2. Define TS interface: `contracts/ts/{module}.ts`
3. Create `modules/{name}/module.yaml`
4. Create `modules/{name}/config.schema.json`
5. Write `modules/{name}/docs/AGENT.md` (agent instructions — be specific)
6. Write `modules/{name}/docs/README.md` (human docs)
7. Create impl stubs: `modules/{name}/impl/{provider}-go/` and `{provider}-ts/`
8. Add contract tests in `modules/{name}/tests/`
9. Add entry to `orchestration/registry.yaml`
10. Update `orchestration/composition-rulebook.md` if new wiring rules are needed

---

## Go Module Path

The contracts Go package: `github.com/dat2503/modkit/contracts/go`
The modkit CLI module: `github.com/dat2503/modkit/modkit`
