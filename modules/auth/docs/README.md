# Auth Module

User authentication and session management for modkit projects.

## Overview

The auth module delegates authentication to a managed provider. The default implementation uses [Clerk](https://clerk.com), which provides:
- Hosted sign-in/sign-up UI components
- JWT token issuance and validation
- User management dashboard
- Webhook events for user lifecycle

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `clerk` | Clerk | MVP | Go, Bun |
| `privy` | Privy | v2 | Go, Bun |

## Setup (Clerk)

1. Create a Clerk application at [dashboard.clerk.com](https://dashboard.clerk.com)
2. Copy your API keys from the dashboard
3. Set env vars:
   ```
   CLERK_SECRET_KEY=sk_test_...
   CLERK_PUBLISHABLE_KEY=pk_test_...
   ```
4. Add the Clerk middleware to your backend (see `impl/clerk-go/` or `impl/clerk-ts/`)
5. Add `@clerk/nextjs` to your frontend

## Frontend integration

This module handles backend token validation only. For frontend auth UI, use the provider's SDK:

```bash
# In apps/web/
bun add @clerk/nextjs
```

Wrap your app with `<ClerkProvider>` and use `<SignIn>`, `<SignUp>`, `<UserButton>` components.

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **cache** (required) — used for session token caching and blacklisting
- **observability** (optional) — traces auth validation calls
