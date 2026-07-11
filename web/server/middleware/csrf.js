// CSRF guard for mutations (TD §2). The session cookie is SameSite=Strict,
// so this is defense-in-depth: reject cross-origin POST/PATCH/PUT/DELETE to
// /api/** whose Origin (or Sec-Fetch-Site) doesn't match this host.
const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

export default defineEventHandler((event) => {
  const path = event.path || event.node.req.url || ''
  if (!path.startsWith('/api/')) return
  if (SAFE_METHODS.has(event.node.req.method)) return

  const secFetchSite = getHeader(event, 'sec-fetch-site')
  if (secFetchSite) {
    if (secFetchSite === 'same-origin' || secFetchSite === 'none') return
    throw createError({
      statusCode: 403,
      data: { error: { code: 'PERMISSION_DENIED', message: 'Cross-site request rejected' } },
    })
  }

  // Sec-Fetch-Site is sent by all modern browsers; fall back to Origin vs
  // Host comparison only for the rare client that omits it.
  const origin = getHeader(event, 'origin')
  if (!origin) return
  const host = getHeader(event, 'host')
  let originHost
  try {
    originHost = new URL(origin).host
  } catch {
    throw createError({
      statusCode: 403,
      data: { error: { code: 'PERMISSION_DENIED', message: 'Invalid Origin header' } },
    })
  }
  if (originHost !== host) {
    throw createError({
      statusCode: 403,
      data: { error: { code: 'PERMISSION_DENIED', message: 'Cross-origin request rejected' } },
    })
  }
})
