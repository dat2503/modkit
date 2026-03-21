# Email Module — Agent Instructions

## When to use

Include this module when:
- App sends transactional emails (welcome, confirmation, password reset)
- App sends notification emails (invoice sent, payment received, status change)
- App sends magic link / OTP emails

Do NOT use for:
- Marketing emails or newsletters — use a dedicated platform (Mailchimp, Loops, etc.)
- Bulk promotional campaigns — different deliverability requirements

## How to wire

### Go

1. Import `EmailService` from `contracts/go/email.go`
2. Initialize in bootstrap:
   ```go
   emailSvc := resend.New(resend.Config{
       APIKey:      cfg.Email.APIKey,
       FromDefault: cfg.Email.FromDefault,
   })
   ```
3. Inject into handlers and job handlers that send email
4. For async sending (recommended for non-critical emails), pair with the jobs module

### Bun (TypeScript)

1. Import `IEmailService` from `contracts/ts/email.ts`
2. Initialize in bootstrap:
   ```typescript
   const email = new ResendEmailService({
     apiKey: config.email.apiKey,
     fromDefault: config.email.fromDefault,
   })
   ```
3. Inject into handlers and job handlers

## Common patterns

### Send a transactional email (sync)

```go
result, err := email.Send(ctx, contracts.EmailMessage{
    To:      []string{"user@example.com"},
    From:    "noreply@yourapp.com",
    Subject: "Your invoice is ready",
    Body: contracts.EmailBody{
        HTML: "<p>Your invoice #123 is ready. <a href='...'>View it here</a>.</p>",
        Text: "Your invoice #123 is ready. View it at: ...",
    },
})
if err != nil {
    // always log email errors — don't silently swallow them
    log.Error("failed to send invoice email", "error", err)
}
```

### Send email asynchronously (preferred for non-critical)

Always pair with the jobs module for emails that don't need to block the response:

```go
// In your handler:
err := jobs.Enqueue(ctx, "email:send_welcome", WelcomeEmailPayload{UserID: user.ID})

// In your job handler:
func handleSendWelcomeEmail(ctx context.Context, payload []byte) error {
    var p WelcomeEmailPayload
    json.Unmarshal(payload, &p)
    user, _ := userRepo.Get(ctx, p.UserID)
    _, err := email.Send(ctx, contracts.EmailMessage{...})
    return err  // return error to trigger retry
}
```

## Email templates

Store HTML templates in `apps/api/templates/email/`. Use Go's `html/template` or
a template library. Never build HTML by string concatenation.

## Domain verification

Your from address domain MUST be verified in the email provider dashboard before
sending. Use a subdomain like `mail.yourapp.com` to protect your main domain's
reputation.

## Required env vars

```
EMAIL_PROVIDER=resend
EMAIL_API_KEY=re_...                # sensitive
EMAIL_FROM_DEFAULT=noreply@yourapp.com
```

## Do NOT

- Store API keys in code — use config only
- Send marketing/bulk email through this module
- Ignore delivery failures — always log errors
- Use a personal email address as the from address
