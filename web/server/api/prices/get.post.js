import { grpcCall } from '~~/server/utils/grpcClient'
import { getPricesClient } from '~~/server/utils/pricesClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { productId } -> ProductPrice | throws NOT_FOUND if the product has no
// price yet (the client treats that as "show the create form", not an
// error toast — see ProductPriceTab.vue).
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getPricesClient()
  try {
    const response = await grpcCall(client, 'Get', { productId: body.productId }, { token: session.token })
    return response.price
  } catch (err) {
    throw mapGrpcError(err)
  }
})
