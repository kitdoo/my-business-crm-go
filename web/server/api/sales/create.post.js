import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { clientId, partnerId?, items: [{skuId, warehouseId, quantity, discountPercentage}] } -> Sale
// clientId is mutually exclusive with:
// { newClient: {name, phone, email, address}, ... } -> Sale
// (find-or-create by email server-side — TD §12.3, no separate "create a
// client first" step on the frontend). Each item carries its own
// warehouseId — a sale can draw different lines from different warehouses.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getServiceClient('sale.proto', 'crm.grpc.sale.v1.SalesService')
  try {
    const response = await grpcCall(
      client,
      'Create',
      {
        ...(body.clientId ? { clientId: body.clientId } : { newClient: body.newClient }),
        partnerId: body.partnerId,
        items: body.items,
      },
      { token: session.token },
    )
    return response.sale
  } catch (err) {
    throw mapGrpcError(err)
  }
})
