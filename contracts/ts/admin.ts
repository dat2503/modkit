import type { AuthUser, UserList, ListUsersOptions } from './auth'

/**
 * AdminService provides admin-only operations: user management and dashboard stats.
 * Requires the auth module — wraps IAuthService for role-based operations.
 */
export interface IAdminService {
  /** Lists all users with pagination. */
  listUsers(opts?: ListUsersOptions): Promise<UserList>

  /** Updates a user's role (e.g. "admin", "user"). */
  setUserRole(userId: string, role: string): Promise<void>

  /** Deletes a user by ID. */
  deleteUser(userId: string): Promise<void>

  /** Returns dashboard statistics. */
  getDashboardStats(): Promise<DashboardStats>

  /** Checks if a user has the admin role. */
  isAdmin(user: AuthUser): boolean
}

export interface DashboardStats {
  totalUsers: number
  recentSignups: number
}
