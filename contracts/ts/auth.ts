/**
 * AuthService handles user authentication and session management.
 * It validates tokens issued by a third-party auth provider (e.g. Clerk).
 * Never implement password storage — always delegate to the provider.
 */
export interface IAuthService {
  /**
   * Validates a JWT or session token and returns the authenticated user.
   * Throws AuthError with code UNAUTHORIZED if the token is invalid or expired.
   */
  validateToken(token: string): Promise<AuthUser>;

  /**
   * Retrieves a user by their provider-assigned ID.
   * Throws AuthError with code NOT_FOUND if the user does not exist.
   */
  getUser(userId: string): Promise<AuthUser>;

  /**
   * Returns a paginated list of users.
   */
  listUsers(opts?: ListUsersOptions): Promise<UserList>;

  /**
   * Removes a user from the auth provider.
   * Call this when a user requests account deletion.
   */
  deleteUser(userId: string): Promise<void>;
}

/** Represents an authenticated user returned from the auth provider. */
export interface AuthUser {
  /** Provider-assigned unique identifier (e.g. "user_2abc123" for Clerk). */
  id: string;

  /** The user's primary email address. */
  email: string;

  /** The user's display name. */
  name: string;

  /** URL of the user's profile picture. May be undefined. */
  avatarUrl?: string;

  /** Arbitrary key-value pairs set on the user in the auth provider. */
  metadata?: Record<string, string>;
}

/** Controls pagination for listUsers. */
export interface ListUsersOptions {
  /** Maximum number of users to return. Defaults to 20, max 100. */
  limit?: number;

  /** Number of users to skip (for page-based pagination). */
  offset?: number;
}

/** Result of a listUsers call. */
export interface UserList {
  /** The users for this page. */
  users: AuthUser[];

  /** Total number of users across all pages. */
  total: number;
}
