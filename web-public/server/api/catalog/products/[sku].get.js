import { getActiveProductBySku, getProductPrice } from '~~/server/utils/catalogClient.js'

// GET /api/catalog/products/[sku] — product detail page (TZ §8.6).
export default defineEventHandler(async (event) => {
  const sku = getRouterParam(event, 'sku')
  const product = await getActiveProductBySku(sku)
  if (!product) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }
  const price = await getProductPrice(product.id)

  return {
    id: product.id,
    sku: product.sku,
    name: product.name,
    description: product.description,
    details: product.details,
    imageIds: product.imageIds,
    price: price ? { amount: price.priceAmount, currency: price.currency } : null,
    status: product.status,
  }
})
