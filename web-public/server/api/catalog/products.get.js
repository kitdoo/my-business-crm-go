import { listActiveProducts, getProductPrice } from '~~/server/utils/catalogClient.js'

// name_asc/name_desc/newest — the only sort options exposed to the public
// site; anything else falls back to newest-first.
const SORT_OPTIONS = {
  name_asc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_ASC' },
  name_desc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_DESC' },
  newest: { sortField: 'FIELD_CREATED_AT', sortDirection: 'SORT_DIRECTION_DESC' },
}

// GET /api/catalog/products?categoryId=&cursor=&limit=&sort= — TZ §8.2.
// Response is trimmed to card fields only, price resolved per item.
export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const sort = SORT_OPTIONS[query.sort?.toString()] || SORT_OPTIONS.newest
  const response = await listActiveProducts({
    categoryId: query.categoryId?.toString(),
    cursor: query.cursor?.toString(),
    limit: query.limit ? Number(query.limit) : undefined,
    ...sort,
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

  return { items, total: response.total ?? null, nextCursor: response.nextCursor || null }
})
