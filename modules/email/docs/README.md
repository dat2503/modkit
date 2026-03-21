# Email Module

Transactional email delivery for modkit projects.

## Overview

The email module sends transactional emails triggered by user actions. The default implementation uses [Resend](https://resend.com), which provides a clean API and excellent deliverability.

## Implementations

| Name | Label | Phase | Runtimes |
|------|-------|-------|---------|
| `resend` | Resend | MVP | Go, Bun |
| `sendgrid` | SendGrid | v2 | Go, Bun |

## Setup (Resend)

1. Create an account at [resend.com](https://resend.com)
2. Add and verify your sending domain
3. Create an API key
4. Set env vars:
   ```
   EMAIL_API_KEY=re_...
   EMAIL_FROM_DEFAULT=noreply@yourdomain.com
   ```

## Configuration

See `config.schema.json` for all environment variables.

## Dependencies

- **jobs** (optional but recommended) — send emails asynchronously with retries
- **observability** (optional) — traces email send calls
