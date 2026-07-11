import { grpcCall } from '~~/server/utils/grpcClient'
import { getPricesClient } from '~~/server/utils/pricesClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { productId, priceAmount, discountAmount? } -> ProductPrice
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getPricesClient()
  try {
    const response = await grpcCall(
      client,
      'Create',
      { productId: body.productId, priceAmount: body.priceAmount, discountAmount: body.discountAmount },
      { token: session.token },
    )
    return response.price
  } catch (err) {
    throw mapGrpcError(err)
  }
})
