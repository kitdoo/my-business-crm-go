import { listActiveProducts, getProductPrice } from '~~/server/utils/catalogClient.js'

// GET /api/catalog/products?categoryId=&cursor=&limit= — TZ §8.2.
// Response is trimmed to card fields only, price resolved per item.
export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const response = await listActiveProducts({
    categoryId: query.categoryId?.toString(),
    cursor: query.cursor?.toString(),
    limit: query.limit ? Number(query.limit) : undefined,
  })

  const items = await Promise.all(
    (response.items || []).map(async (product) => {
      const price = await getProductPrice(product.id)
      return {
        id: product.id,
        sku: product.sku,
        name: product.name,
        imageIds: product.imageIds,
        price: price ? { amount: price.priceAmount, currency: price.currency } : null,
        status: product.status,
      }
    }),
  )

  return { items, nextCursor: response.nextCursor || null }
})
