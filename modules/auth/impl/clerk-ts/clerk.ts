import type { IAuthService, AuthUser, ListUsersOptions, UserList } from '../../../contracts/ts/auth'
import type { ICacheService } from '../../../contracts/ts/cache'

export interface ClerkConfig {
  secretKey: string
  publishableKey: string
}

/**
 * ClerkAuthService implements IAuthService using Clerk (clerk.com).
 */
export class ClerkAuthService implements IAuthService {
  constructor(
    private readonly config: ClerkConfig,
    private readonly cache: ICacheService,
  ) {}

  async validateToken(token: string): Promise<AuthUser> {
    // TODO: implement using @clerk/backend
    // 1. verifyToken(token, { secretKey: this.config.secretKey })
    // 2. Extract userId from claims
    // 3. Optionally cache validated user
    console.warn('[clerk-auth] stub: validateToken() not implemented')
    return { id: '', email: '', name: '' }
  }

  async getUser(userId: string): Promise<AuthUser> {
    // TODO: implement using @clerk/backend clerkClient.users.getUser(userId)
    console.warn('[clerk-auth] stub: getUser() not implemented')
    return { id: userId, email: '', name: '' }
  }

  async listUsers(opts?: ListUsersOptions): Promise<UserList> {
    // TODO: implement using @clerk/backend clerkClient.users.getUserList()
    console.warn('[clerk-auth] stub: listUsers() not implemented')
    return { users: [], total: 0 }
  }

  async deleteUser(userId: string): Promise<void> {
    // TODO: implement using @clerk/backend clerkClient.users.deleteUser(userId)
    console.warn('[clerk-auth] stub: deleteUser() not implemented')
  }

  async updateUserRole(userId: string, role: string): Promise<void> {
    // TODO: implement using @clerk/backend clerkClient.users.updateUserMetadata(userId, { publicMetadata: { role } })
    console.warn('[clerk-auth] stub: updateUserRole() not implemented')
  }
}
