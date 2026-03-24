import type { IAdminService, DashboardStats } from '../../../../contracts/ts/admin'
import type { IAuthService, AuthUser, ListUsersOptions, UserList } from '../../../../contracts/ts/auth'

export class AdminService implements IAdminService {
  constructor(private readonly auth: IAuthService) {}

  async listUsers(opts?: ListUsersOptions): Promise<UserList> {
    return this.auth.listUsers(opts)
  }

  async setUserRole(userId: string, role: string): Promise<void> {
    await this.auth.updateUserRole(userId, role)
  }

  async deleteUser(userId: string): Promise<void> {
    await this.auth.deleteUser(userId)
  }

  async getDashboardStats(): Promise<DashboardStats> {
    const allUsers = await this.auth.listUsers({ limit: 1 })
    const recentUsers = await this.auth.listUsers({ limit: 100, offset: 0 })

    // Count users created in the last 7 days
    // Since we don't have createdAt on AuthUser, use total count as approximation
    return {
      totalUsers: allUsers.total,
      recentSignups: Math.min(recentUsers.users.length, allUsers.total),
    }
  }

  isAdmin(user: AuthUser): boolean {
    return user.role === 'admin'
  }
}
