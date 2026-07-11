import { randomUUID } from 'node:crypto'

// Server-side session store (TD §2): the gRPC token from UsersService.Login
// never leaves this process. The browser only ever holds an opaque
// httpOnly cookie with a session id that indexes into this map.
//
// In-memory Map today; swap for a Redis-backed store (configs/redis.yaml)
// behind the same get/set/destroy interface once the app runs multi-instance
// — the cookie/session-id contract on the browser side does not change.
const sessions = new Map()

const SESSION_COOKIE = 'crm_session'

function now() {
  return Date.now()
}

function pruneExpired() {
  for (const [id, session] of sessions) {
    if (session.expiresAt <= now()) sessions.delete(id)
  }
}

/**
 * @param {import('h3').H3Event} event
 * @param {{ token: string, user: object }} data
 */
export function createSession(event, data) {
  pruneExpired()
  const config = useRuntimeConfig()
  const id = randomUUID()
  const ttlMs = (config.session.ttlSeconds || 60 * 60 * 12) * 1000
  sessions.set(id, { token: data.token, user: data.user, expiresAt: now() + ttlMs })

  setCookie(event, SESSION_COOKIE, id, {
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    path: '/',
    maxAge: Math.floor(ttlMs / 1000),
  })
  return id
}

/**
 * Named getCrmSession (not getSession) to avoid colliding with h3's
 * built-in sealed-cookie session utility of the same name — we deliberately
 * don't use that one, since it would put the token in a client-readable
 * (if encrypted) cookie instead of keeping it server-only (TD §2).
 * @param {import('h3').H3Event} event
 * @returns {{ token: string, user: object } | null}
 */
export function getCrmSession(event) {
  const id = getCookie(event, SESSION_COOKIE)
  if (!id) return null
  const session = sessions.get(id)
  if (!session || session.expiresAt <= now()) {
    sessions.delete(id)
    return null
  }
  return session
}

/**
 * @param {import('h3').H3Event} event
 */
export function destroySession(event) {
  const id = getCookie(event, SESSION_COOKIE)
  if (id) sessions.delete(id)
  deleteCookie(event, SESSION_COOKIE, { path: '/' })
}

/**
 * Throws a 401 H3Error (mapped shape matches mapGrpcError's UNAUTHENTICATED)
 * if there is no valid session. Use at the top of every authenticated
 * server route.
 * @param {import('h3').H3Event} event
 */
export function requireSession(event) {
  const session = getCrmSession(event)
  if (!session) {
    throw createError({
      statusCode: 401,
      data: { error: { code: 'UNAUTHENTICATED', message: 'Session expired or missing' } },
    })
  }
  return session
}
