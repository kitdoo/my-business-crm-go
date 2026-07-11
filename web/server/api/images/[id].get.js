// Proxies GET /images/{id} on the backend HTTP host (TD §4.6) — unauthenticated
// on the backend (catalog images are public), proxied anyway so the browser
// never learns the real backend host and everything stays same-origin.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = getRouterParam(event, 'id')

  let response
  try {
    response = await $fetch.raw(`${config.images.baseUrl}/images/${id}`, { responseType: 'arrayBuffer' })
  } catch {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  setHeader(event, 'Content-Type', response.headers.get('content-type') || 'application/octet-stream')
  setHeader(event, 'Cache-Control', 'public, max-age=31536000, immutable')
  return Buffer.from(response._data)
})
