import { destroySession } from '~~/server/utils/session'

// "Logout" only destroys the local session — the backend token itself has
// no revocation RPC and stays valid until its TTL expires (TD §4.1 step 4).
export default defineEventHandler((event) => {
  destroySession(event)
  return { ok: true }
})
