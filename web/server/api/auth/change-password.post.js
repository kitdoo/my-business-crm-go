import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// Self-service, not part of the Users entity CRUD routes — any
// authenticated caller may change their own password (TD §4.2/§12.5),
// gated only by knowing the current password, not RBAC. The target id is
// always the caller's own session user, never taken from the request body.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getServiceClient('user.proto', 'crm.grpc.user.v1.UsersService')
  try {
    await grpcCall(
      client,
      'ChangePassword',
      { id: session.user.id, currentPassword: body.currentPassword, newPassword: body.newPassword },
      { token: session.token },
    )
    return { ok: true }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
