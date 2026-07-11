import { grpcCall } from '~~/server/utils/grpcClient'
import { getReportsClient } from '~~/server/utils/reportsClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const body = (await readBody(event).catch(() => ({}))) || {}
  const client = getReportsClient()
  try {
    const response = await grpcCall(
      client,
      'GetPopularProducts',
      { period: body.period, limit: body.limit || 10 },
      { token: session.token },
    )
    return { rows: response.rows || [] }
  } catch (err) {
    throw mapGrpcError(err)
  }
})
