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
    throw new Error('not implemented')
  }

  async getUser(userId: string): Promise<AuthUser> {
    // TODO: implement using @clerk/backend clerkClient.users.getUser(userId)
    throw new Error('not implemented')
  }

  async listUsers(opts?: ListUsersOptions): Promise<UserList> {
    // TODO: implement using @clerk/backend clerkClient.users.getUserList()
    throw new Error('not implemented')
  }

  async deleteUser(userId: string): Promise<void> {
    // TODO: implement using @clerk/backend clerkClient.users.deleteUser(userId)
    throw new Error('not implemented')
  }

  async updateUserRole(userId: string, role: string): Promise<void> {
    // TODO: implement using @clerk/backend clerkClient.users.updateUserMetadata(userId, { publicMetadata: { role } })
    throw new Error('not implemented')
  }
}
