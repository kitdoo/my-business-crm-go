import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// Not part of the standard 5 CRUD routes (defineEntityHandler) — a
// dedicated RPC per TD §12.2: WarehouseUpdateRequest doesn't even carry a
// status field, so deactivating a warehouse can't go through Update.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getServiceClient('warehouse.proto', 'crm.grpc.warehouse.v1.WarehousesService')
  try {
    const response = await grpcCall(
      client,
      'Deactivate',
      { id: body.id, etag: body.etag },
      { token: session.token },
    )
    return response.warehouse
  } catch (err) {
    throw mapGrpcError(err)
  }
})
