import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { createSession } from '~~/server/utils/session'
import { checkRateLimit } from '~~/server/utils/rateLimiter'
import { publicUser } from '~~/server/utils/publicUser'

export default defineEventHandler(async (event) => {
  const ip = getRequestIP(event, { xForwardedFor: true }) || 'unknown'
  if (!checkRateLimit(`login:${ip}`, { max: 10, windowMs: 60_000 })) {
    throw createError({
      statusCode: 429,
      data: { error: { code: 'RESOURCE_EXHAUSTED', message: 'Too many login attempts, try again shortly' } },
    })
  }

  const body = await readBody(event)
  const login = typeof body?.login === 'string' ? body.login : ''
  const password = typeof body?.password === 'string' ? body.password : ''
  if (!login || !password) {
    throw createError({
      statusCode: 400,
      data: { error: { code: 'INVALID_ARGUMENT', message: 'login and password are required' } },
    })
  }

  const client = getServiceClient('user.proto', 'crm.grpc.user.v1.UsersService')
  let response
  try {
    response = await grpcCall(client, 'Login', { login, password })
  } catch (err) {
    throw mapGrpcError(err)
  }

  createSession(event, { token: response.token, user: response.user })
  return { user: publicUser(response.user) }
})
