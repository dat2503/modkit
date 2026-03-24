import type { IAuthService, AuthUser, ListUsersOptions, UserList } from '../../../contracts/ts/auth'
import type { ICacheService } from '../../../contracts/ts/cache'

export interface BetterAuthConfig {
  /** Base URL where the Better Auth server is running, e.g. "http://localhost:3000" */
  baseUrl: string
  /** The BETTER_AUTH_SECRET — must match the server's secret */
  secret: string
}

/**
 * BetterAuthService implements IAuthService using Better Auth (https://www.better-auth.com).
 *
 * For the Bun runtime, Better Auth runs embedded in the same process.
 * This service calls Better Auth's admin API endpoints to validate sessions and manage users.
 */
export class BetterAuthService implements IAuthService {
  constructor(
    private readonly config: BetterAuthConfig,
    private readonly cache: ICacheService,
  ) {}

  async validateToken(token: string): Promise<AuthUser> {
    const res = await fetch(`${this.config.baseUrl}/api/auth/get-session`, {
      headers: { Authorization: `Bearer ${token}` },
    })

    if (res.status === 401) {
      throw new Error('better-auth: invalid or expired token')
    }
    if (!res.ok) {
      throw new Error(`better-auth: get-session returned ${res.status}`)
    }

    const data = await res.json() as { user?: { id: string; email: string; name: string; image?: string; role?: string } }
    if (!data.user?.id) {
      throw new Error('better-auth: session has no user')
    }

    return {
      id: data.user.id,
      email: data.user.email,
      name: data.user.name,
      avatarUrl: data.user.image ?? '',
      role: data.user.role ?? 'user',
      metadata: {},
    }
  }

  async getUser(userId: string): Promise<AuthUser> {
    const url = `${this.config.baseUrl}/api/auth/admin/list-users?searchField=id&searchValue=${encodeURIComponent(userId)}&limit=1`
    const res = await fetch(url, {
      headers: { 'x-better-auth-secret': this.config.secret },
    })

    if (!res.ok) {
      throw new Error(`better-auth: get user returned ${res.status}`)
    }

    const data = await res.json() as { users: BAUser[] }
    if (!data.users?.length) {
      throw new Error(`better-auth: user "${userId}" not found`)
    }

    return baUserToContract(data.users[0])
  }

  async listUsers(opts?: ListUsersOptions): Promise<UserList> {
    const limit = opts?.limit ?? 20
    const offset = opts?.offset ?? 0
    const url = `${this.config.baseUrl}/api/auth/admin/list-users?limit=${limit}&offset=${offset}`
    const res = await fetch(url, {
      headers: { 'x-better-auth-secret': this.config.secret },
    })

    if (!res.ok) {
      throw new Error(`better-auth: list users returned ${res.status}`)
    }

    const data = await res.json() as { users: BAUser[]; total: number }
    return {
      users: (data.users ?? []).map(baUserToContract),
      total: data.total ?? 0,
    }
  }

  async updateUserRole(userId: string, role: string): Promise<void> {
    const res = await fetch(`${this.config.baseUrl}/api/auth/admin/set-role`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-better-auth-secret': this.config.secret,
      },
      body: JSON.stringify({ userId, role }),
    })

    if (!res.ok) {
      throw new Error(`better-auth: set role returned ${res.status}`)
    }
  }

  async deleteUser(userId: string): Promise<void> {
    const res = await fetch(`${this.config.baseUrl}/api/auth/admin/remove-user`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        'x-better-auth-secret': this.config.secret,
      },
      body: JSON.stringify({ userId }),
    })

    if (!res.ok && res.status !== 204) {
      throw new Error(`better-auth: delete user returned ${res.status}`)
    }
  }
}

interface BAUser {
  id: string
  email: string
  name: string
  image?: string
  role?: string
}

function baUserToContract(u: BAUser): AuthUser {
  return {
    id: u.id,
    email: u.email,
    name: u.name,
    avatarUrl: u.image ?? '',
    role: u.role ?? 'user',
    metadata: {},
  }
}
