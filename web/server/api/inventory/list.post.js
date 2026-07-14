import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// Not defineEntityHandler: InventoryListRequest.variantId/warehouseId are
// top-level fields, not nested under Filter like every other entity's
// list request — so EntityDataTable's fixedFilter (which always merges
// into `filter`) needs pulling apart here before the gRPC call. Also
// read-only on the backend (Get/List only, TD §10), so this is the only
// route this entity has.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const { variantId, warehouseId, ...filter } = body.filter || {}
  const client = getServiceClient('inventory.proto', 'crm.grpc.inventory.v1.InventoryService')
  try {
    const response = await grpcCall(
      client,
      'List',
      { variantId, warehouseId, filter, pagination: body.pagination },
      { token: session.token },
    )
    return { items: response.items || [], nextCursor: response.nextCursor || null }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
