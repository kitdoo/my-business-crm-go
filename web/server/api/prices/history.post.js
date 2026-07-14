import { grpcCall } from '~~/server/utils/grpcClient'
import { getPricesClient } from '~~/server/utils/pricesClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { variantId, filter?, pagination? } -> { items, nextCursor }
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getPricesClient()
  try {
    const response = await grpcCall(
      client,
      'GetHistory',
      { variantId: body.variantId, filter: body.filter, pagination: body.pagination },
      { token: session.token },
    )
    return { items: response.items || [], nextCursor: response.nextCursor || null }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
