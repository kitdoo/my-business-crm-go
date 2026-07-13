import {
  getActiveProductBySku,
  getProductPrice,
  listPublicAttributeDefinitions,
} from '~~/server/utils/catalogClient.js'

// GET /api/catalog/products/[sku] — product detail page (TZ §8.6). Details
// is filtered down to only publicly-visible characteristic keys here —
// never trust the backend's raw map, private keys must not reach a site
// visitor (see listPublicAttributeDefinitions).
export default defineEventHandler(async (event) => {
  const sku = getRouterParam(event, 'sku')
  const product = await getActiveProductBySku(sku)
  if (!product) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }
  const [price, publicDefinitions] = await Promise.all([
    getProductPrice(product.id),
    listPublicAttributeDefinitions(),
  ])
  const publicKeys = new Set(publicDefinitions.map((d) => d.key))
  const details = Object.fromEntries(Object.entries(product.details || {}).filter(([key]) => publicKeys.has(key)))

  return {
    id: product.id,
    sku: product.sku,
    name: product.name,
    description: product.description,
    details,
    imageIds: product.imageIds,
    price: price ? { amount: price.priceAmount, currency: price.currency } : null,
    status: product.status,
  }
})
