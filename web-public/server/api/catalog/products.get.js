import {
  listActiveProducts,
  listActiveVariantsForProduct,
  getVariantPrice,
  isVariantInStock,
} from '~~/server/utils/catalogClient.js'

// name_asc/name_desc/newest — the only sort options exposed to the public
// site; anything else falls back to newest-first.
const SORT_OPTIONS = {
  name_asc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_ASC' },
  name_desc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_DESC' },
  newest: { sortField: 'FIELD_CREATED_AT', sortDirection: 'SORT_DIRECTION_DESC' },
}

// A product card shows one "representative" variant — its first active
// variant (list is already sorted by creation order) — since price/image
// vary per variant, not per product. A product with no active variant
// yet (still being set up) is dropped from the catalog entirely.
async function toCard(product) {
  const variants = await listActiveVariantsForProduct(product.id)
  const variant = variants[0]
  if (!variant) return null

  const [price, inStock] = await Promise.all([getVariantPrice(variant.id), isVariantInStock(variant.id)])
  return {
    id: product.id,
    sku: variant.sku,
    name: product.name,
    imageIds: variant.imageIds,
    price: price ? { amount: price.priceAmount, currency: price.currency } : null,
    status: product.status,
    inStock,
  }
}

// InventoryService has no "has stock" filter of its own (TZ §8.2 addendum:
// stock filter buttons on /katalog) — ProductsService's cursor pages are
// pulled and checked against Inventory one raw page at a time until enough
// matching cards are collected or the backend runs out of pages. The
// iteration cap bounds worst case (e.g. every item on a page filtered out)
// to a handful of extra round trips instead of scanning the whole catalog.
const MAX_RAW_PAGES = 15

// Always consumes whole raw pages (never slices mid-page) so nextCursor —
// an opaque backend cursor, not an item offset — stays valid for the next
// call; a returned page can therefore run a little over `limit`.
async function loadFilteredPage({ categoryId, cursor, limit, sortField, sortDirection, wantInStock }) {
  const items = []
  let nextCursor = cursor
  for (let i = 0; i < MAX_RAW_PAGES && items.length < limit; i++) {
    const response = await listActiveProducts({ categoryId, cursor: nextCursor, limit, sortField, sortDirection })
    const rawItems = response.items || []
    nextCursor = response.nextCursor || null

    const cards = (await Promise.all(rawItems.map(toCard))).filter(Boolean)
    for (const card of cards) {
      if (card.inStock === wantInStock) items.push(card)
    }

    if (!nextCursor) break
  }
  return { items, nextCursor }
}

// GET /api/catalog/products?categoryId=&cursor=&limit=&sort=&inStock= — TZ §8.2.
// Response is trimmed to card fields only, sourced from each product's
// representative variant (see toCard). inStock=true|false filters by
// warehouse stock (see catalogClient.isVariantInStock); omitted means no
// stock filter. total is only reliable when unfiltered — a stock filter
// makes ProductsService's count meaningless, so it's dropped.
export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const sort = SORT_OPTIONS[query.sort?.toString()] || SORT_OPTIONS.newest
  const categoryId = query.categoryId?.toString()
  const cursor = query.cursor?.toString()
  const limit = query.limit ? Number(query.limit) : 12
  const inStockParam = query.inStock?.toString()

  if (inStockParam === 'true' || inStockParam === 'false') {
    const { items, nextCursor } = await loadFilteredPage({
      categoryId,
      cursor,
      limit,
      ...sort,
      wantInStock: inStockParam === 'true',
    })
    return { items, total: null, nextCursor }
  }

  const response = await listActiveProducts({ categoryId, cursor, limit, ...sort })
  const items = (await Promise.all((response.items || []).map(toCard))).filter(Boolean)
  return { items, total: response.total ?? null, nextCursor: response.nextCursor || null }
})
