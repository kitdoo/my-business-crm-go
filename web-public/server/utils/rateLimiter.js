// Simple in-memory fixed-window rate limiter (TD §4.5) — the backend does
// not rate-limit Login itself, so the BFF adds a minimal brute-force guard.
const buckets = new Map()

/**
 * @param {string} key e.g. `login:${clientIp}`
 * @param {{ max: number, windowMs: number }} opts
 * @returns {boolean} true if the call is allowed
 */
export function checkRateLimit(key, { max, windowMs }) {
  const now = Date.now()
  const bucket = buckets.get(key)
  if (!bucket || bucket.resetAt <= now) {
    buckets.set(key, { count: 1, resetAt: now + windowMs })
    return true
  }
  if (bucket.count >= max) return false
  bucket.count += 1
  return true
}
