import { grpcCall } from '~~/server/utils/grpcClient'
import { getPricesClient } from '~~/server/utils/pricesClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

// { id, fields, mask, etag } -> ProductPrice — same wire shape as the
// generic useEntityApi().update() so ProductPriceTab.vue can reuse
// buildUpdateMask the same way every other form does.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getPricesClient()
  try {
    const response = await grpcCall(
      client,
      'Update',
      {
        id: body.id,
        ...body.fields,
        options: { updateMask: { paths: body.mask || [] }, etag: body.etag },
      },
      { token: session.token },
    )
    return response.price
  } catch (err) {
    throw mapGrpcError(err)
  }
})
