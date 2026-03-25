# Formly — Form Builder SaaS

## The idea

I want to build a SaaS tool for freelancers and small businesses to create custom forms, share them publicly, and get notified when someone submits a response.

## Who it's for

- Freelancers collecting client briefs
- Small agencies running lead capture forms
- Anyone who needs a quick form without paying $30/mo for Typeform

## What it does

**For the form owner (signed-in user):**
- Sign up and log in
- Create forms with a title and a list of fields (short text, long text, multiple choice, file upload)
- Each form gets a unique public link to share
- See all submissions in a dashboard
- Get an email notification every time someone submits their form
- Upgrade to a Pro plan to unlock more than 3 forms and file upload fields

**For the person filling out the form (public, no login):**
- Open the public form link
- Fill it out and submit
- See a confirmation message

**Admin (internal):**
- View all users and their plan status
- Manually upgrade or downgrade a user's plan

## Tech preferences

- Backend in Go, frontend in React (Vite)
- Postgres for the database
- Keep it simple — no over-engineering

## What success looks like

A user can sign up, create a form, share the link, receive a submission, get an email about it, and upgrade to Pro via a Stripe checkout. That full loop working end to end is the goal.
