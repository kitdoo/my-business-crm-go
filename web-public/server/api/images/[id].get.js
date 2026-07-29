// Copied from web/server/api/images/[id].get.js (admin) — GET /images/{id}
// is unauthenticated on the backend HTTP host (catalog images are public),
// proxied anyway so the browser never learns the real backend host.
//
// Streams the response via proxyRequest instead of buffering the whole
// image into memory (the previous $fetch.raw + arrayBuffer approach read
// every image fully into Node before re-sending it, which added latency
// and memory pressure once several images loaded concurrently on one
// page). ?w= is forwarded through to the backend, which resizes to one of
// a fixed set of widths (see internal/transports/http/handlers/image) —
// callers pass it for grid/thumbnail images to avoid shipping full-size
// originals just to display them small.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')
  const { w } = getQuery(event)

  const target = new URL(`${config.images.baseUrl}/images/${id}`)
  if (w) target.searchParams.set('w', w)

  try {
    return await proxyRequest(event, target.toString())
  } catch {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }
})
