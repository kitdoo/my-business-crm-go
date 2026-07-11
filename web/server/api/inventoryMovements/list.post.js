import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// Not defineEntityHandler: InventoryMovementsListRequest.warehouseId is a
// top-level field, not nested under Filter — same shape mismatch as
// Inventory's list. Also append-only (Create + List/GetHistory only, no
// Get/Update/Delete at all, TD §10/§12.4).
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const { warehouseId, ...filter } = body.filter || {}
  const client = getServiceClient('inventory_movement.proto', 'crm.grpc.inventory_movement.v1.InventoryMovementsService')
  try {
    const response = await grpcCall(client, 'List', { warehouseId, filter, pagination: body.pagination }, { token: session.token })
    return { items: response.items || [], nextCursor: response.nextCursor || null }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
