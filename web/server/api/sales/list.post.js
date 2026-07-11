import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// Not defineEntityHandler: SalesListRequest.clientId/warehouseId/partnerId
// are top-level fields, not nested under Filter — same shape as
// Inventory/InventoryMovements. Sale otherwise has no generic
// Update/Delete (TD §12.3), so this file plus get/create/update-status/
// cancel below are all hand-written.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const { clientId, warehouseId, partnerId, ...filter } = body.filter || {}
  const client = getServiceClient('sale.proto', 'crm.grpc.sale.v1.SalesService')
  try {
    const response = await grpcCall(
      client,
      'List',
      {
        clientId,
        warehouseId,
        partnerId,
        sort: body.sort,
        filter,
        pagination: body.pagination,
        options: { includeTotalCount: true },
      },
      { token: session.token },
    )
    return { items: response.items || [], total: response.total, nextCursor: response.nextCursor || null }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
