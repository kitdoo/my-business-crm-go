import { requireSession } from '~~/server/utils/session'

// Proxies POST /invoices/{saleId} on the backend HTTP host (plain HTTP, not
// gRPC — same reasoning as web/server/api/images/upload.post.js, and the
// same backend HTTP server, hence reusing config.images.baseUrl rather than
// adding a second identical baseUrl setting) — admin/employee-only there
// (checked via the same bearer token as gRPC), so the browser never gets a
// direct route to it: same-origin only.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const config = useRuntimeConfig()
  const saleId = getRouterParam(event, 'saleId')
  const body = await readBody(event)

  try {
    const pdf = await $fetch(`${config.images.baseUrl}/invoices/${saleId}`, {
      method: 'POST',
      headers: { authorization: `Bearer ${session.token}` },
      body,
      responseType: 'arrayBuffer',
    })
    setResponseHeader(event, 'Content-Type', 'application/pdf')
    setResponseHeader(event, 'Content-Disposition', 'attachment; filename="invoice.pdf"')
    return Buffer.from(pdf)
  } catch (err) {
    const status = err?.response?.status || err?.statusCode || 502
    const CODE_BY_STATUS = {
      400: 'INVALID_ARGUMENT',
      401: 'UNAUTHENTICATED',
      403: 'PERMISSION_DENIED',
      404: 'NOT_FOUND',
      422: 'FAILED_PRECONDITION',
      503: 'UNAVAILABLE',
    }
    throw createError({
      statusCode: status,
      data: { error: { code: CODE_BY_STATUS[status] || 'INTERNAL', message: 'Invoice generation failed' } },
    })
  }
})
