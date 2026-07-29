// Proxies GET /images/{id} on the backend HTTP host (TD §4.6) — unauthenticated
// on the backend (catalog images are public), proxied anyway so the browser
// never learns the real backend host and everything stays same-origin.
//
// Streams the response via proxyRequest instead of buffering the whole
// image into memory — see web-public/server/api/images/[id].get.js for why.
// ?w= is forwarded through to the backend for resized thumbnails (e.g.
// ImageThumbnail.vue's table-cell previews).
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
