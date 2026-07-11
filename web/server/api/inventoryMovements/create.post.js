import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getServiceClient('inventory_movement.proto', 'crm.grpc.inventory_movement.v1.InventoryMovementsService')
  try {
    const response = await grpcCall(client, 'Create', { ...body.fields }, { token: session.token })
    return response.movement
  } catch (err) {
    throw mapGrpcError(err)
  }
})
