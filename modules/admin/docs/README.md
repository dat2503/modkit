# Admin Module

Adds an admin dashboard with user management and role-based access control.

## Features
- Admin-only route protection via middleware
- User listing with role management
- Dashboard with basic stats (user count, recent signups)
- Seed script to create the initial admin account

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `ADMIN_DEFAULT_EMAIL` | Email for the seeded admin account | `admin@localhost` |
| `ADMIN_DEFAULT_PASSWORD` | Password for the seeded admin account | `changeme` |

## Setup

1. Add the admin module during `modkit init` or via `modkit add-module admin`
2. Run the seed script to create the default admin:
   - **Bun:** `bun run seed`
   - **Go:** `go run cmd/seed/main.go`
3. Sign in with the admin credentials and access `/admin`

## Architecture

The admin module depends on the **auth** module. It uses the `updateUserRole` method on the auth service to assign roles, and checks `user.role` in middleware to gate admin routes.

Admin routes are mounted under `/api/admin/` and require an authenticated user with `role: "admin"`.
