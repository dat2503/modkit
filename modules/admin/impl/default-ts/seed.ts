/**
 * Seed script — creates the default admin account.
 *
 * Usage:
 *   bun run seed
 *
 * Reads ADMIN_DEFAULT_EMAIL and ADMIN_DEFAULT_PASSWORD from .env,
 * creates the account via Better Auth signup, then sets role to "admin".
 */

const email = process.env.ADMIN_DEFAULT_EMAIL ?? 'admin@localhost'
const password = process.env.ADMIN_DEFAULT_PASSWORD ?? 'changeme'
const authUrl = process.env.BETTER_AUTH_URL ?? 'http://localhost:8080'
const secret = process.env.BETTER_AUTH_SECRET ?? ''

async function seed() {
  console.log(`Creating admin account: ${email}`)

  // 1. Sign up the admin user
  const signupRes = await fetch(`${authUrl}/api/auth/sign-up/email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Admin', email, password }),
  })

  if (!signupRes.ok && signupRes.status !== 409) {
    const body = await signupRes.text()
    console.error(`Failed to create admin: ${signupRes.status} ${body}`)
    process.exit(1)
  }

  if (signupRes.status === 409) {
    console.log('Admin account already exists, updating role...')
  }

  // 2. Get the user ID
  const listRes = await fetch(
    `${authUrl}/api/auth/admin/list-users?searchField=email&searchValue=${encodeURIComponent(email)}&limit=1`,
    { headers: { 'x-better-auth-secret': secret } },
  )
  const listData = (await listRes.json()) as { users: Array<{ id: string }> }
  const userId = listData.users?.[0]?.id
  if (!userId) {
    console.error('Could not find admin user after signup')
    process.exit(1)
  }

  // 3. Set role to admin
  const roleRes = await fetch(`${authUrl}/api/auth/admin/set-role`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'x-better-auth-secret': secret,
    },
    body: JSON.stringify({ userId, role: 'admin' }),
  })

  if (!roleRes.ok) {
    console.error(`Failed to set admin role: ${roleRes.status}`)
    process.exit(1)
  }

  console.log(`Admin account ready: ${email} (role: admin)`)
}

seed().catch((err) => {
  console.error(err)
  process.exit(1)
})
