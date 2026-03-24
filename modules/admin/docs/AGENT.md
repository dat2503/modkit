# Admin Module — Agent Instructions

## When to use
Select this module when the project needs:
- An admin dashboard with user management
- Role-based access control (RBAC)
- A default admin account for local development

## Dependencies
- **auth** (required) — admin module wraps the auth service for role checks and user management

## How to wire

### TypeScript (Bun)
```ts
import { AdminService } from '../modules/admin/admin'

const admin = new AdminService(auth)
```

### Go
```go
import adminpkg "yourproject/modules/admin"

adminSvc := adminpkg.New(authSvc)
```

## Init order
Admin initializes **after auth** (step 5 — remaining modules).

## Common patterns

### Admin middleware
The admin module provides middleware that checks `user.role === "admin"`:
- TS: `adminRequired` middleware for Hono
- Go: `AdminRequired` middleware for net/http

### Seed script
Run the seed script to create the default admin account:
- Bun: `bun run seed`
- Go: `go run cmd/seed/main.go`

The seed script uses `ADMIN_DEFAULT_EMAIL` and `ADMIN_DEFAULT_PASSWORD` env vars.

## Do-nots
- Do NOT hardcode admin emails — use the seed script and role field
- Do NOT bypass auth middleware for admin routes — always validate the session first, then check the role
- Do NOT store admin status in a separate table — use the auth provider's role field
