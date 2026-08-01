import { requireSession } from '~~/server/utils/session'

// Proxies POST /sales-report on the backend HTTP host (plain HTTP, not
// gRPC — same reasoning as web/server/api/invoices/[saleId].post.js, and
// the same backend HTTP server, hence reusing config.images.baseUrl rather
// than adding a second identical baseUrl setting) — admin/employee-only
// there (checked via the same bearer token as gRPC), so the browser never
// gets a direct route to it: same-origin only.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const config = useRuntimeConfig()
  const body = await readBody(event)

  try {
    const xlsx = await $fetch(`${config.images.baseUrl}/sales-report`, {
      method: 'POST',
      headers: { authorization: `Bearer ${session.token}` },
      body,
      responseType: 'arrayBuffer',
    })
    setResponseHeader(event, 'Content-Type', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
    setResponseHeader(event, 'Content-Disposition', 'attachment; filename="sales-report.xlsx"')
    return Buffer.from(xlsx)
  } catch (err) {
    const status = err?.response?.status || err?.statusCode || 502
    const CODE_BY_STATUS = {
      400: 'INVALID_ARGUMENT',
      401: 'UNAUTHENTICATED',
      403: 'PERMISSION_DENIED',
    }
    throw createError({
      statusCode: status,
      data: { error: { code: CODE_BY_STATUS[status] || 'INTERNAL', message: 'Sales report generation failed' } },
    })
  }
})
