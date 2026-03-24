import type { Context, Next } from 'hono'
import type { IAuthService } from '../../../../contracts/ts/auth'

/**
 * adminRequired is Hono middleware that ensures the request is from an admin user.
 * Must be used AFTER auth middleware that sets the session token.
 */
export function adminRequired(auth: IAuthService) {
  return async (c: Context, next: Next) => {
    const token = c.req.header('Authorization')?.replace('Bearer ', '')
    if (!token) {
      return c.json({ error: 'Unauthorized' }, 401)
    }

    const user = await auth.validateToken(token)
    if (user.role !== 'admin') {
      return c.json({ error: 'Forbidden: admin access required' }, 403)
    }

    c.set('adminUser', user)
    await next()
  }
}
