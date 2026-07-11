import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { clientId, warehouseId, partnerId?, items: [{productId, quantity, discountPercentage}] } -> Sale
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getServiceClient('sale.proto', 'crm.grpc.sale.v1.SalesService')
  try {
    const response = await grpcCall(
      client,
      'Create',
      {
        clientId: body.clientId,
        warehouseId: body.warehouseId,
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
