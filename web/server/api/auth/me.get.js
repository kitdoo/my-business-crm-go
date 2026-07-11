import { getCrmSession } from '~~/server/utils/session'
import { publicUser } from '~~/server/utils/publicUser'

export default defineEventHandler((event) => {
  const session = getCrmSession(event)
  if (!session) return { user: null }
  return { user: publicUser(session.user) }
})
