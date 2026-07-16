import {
  listActiveProducts,
  listActiveVariantsForProduct,
  listActiveSkusForVariant,
  getSkuPrice,
  isSkuInStock,
  listPublicAttributeDefinitions,
} from '~~/server/utils/catalogClient.js'

function filterPublic(details, publicKeys) {
  return Object.fromEntries(Object.entries(details || {}).filter(([key]) => publicKeys.has(key)))
}

// in_stock/name_asc/name_desc/newest — the only sort options exposed to
// the public site; anything else falls back to in_stock-first. in_stock
// sorts on Product.HasStock, a field ProductsService maintains itself
// (see internal/services/inventory/inventory/service.go's
// recomputeProductHasStock) — not derived here like the old inStock
// filter's page-scan.
const SORT_OPTIONS = {
  in_stock: { sortField: 'FIELD_IN_STOCK', sortDirection: 'SORT_DIRECTION_DESC' },
  name_asc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_ASC' },
  name_desc: { sortField: 'FIELD_NAME', sortDirection: 'SORT_DIRECTION_DESC' },
  newest: { sortField: 'FIELD_CREATED_AT', sortDirection: 'SORT_DIRECTION_DESC' },
}

// A product card shows one "representative" variant, and within it one
// "representative" SKU — the first active one of each (lists are already
// sorted by creation order) — for the card's own price/image/stock, since
// those vary per SKU, not per product or even per variant. A product with
// no active variant, or a variant with no active SKU, is dropped from the
// catalog entirely (nothing purchasable to show). Every sibling variant
// and every sibling SKU within it gets its own price/stock too (not just
// the representative) so the card's two-tier option pickers (variant
// swatches, then SKU pills) can swap the preview photo/price/availability
// in place instead of only swapping the photo.
async function toCard(product, publicKeys) {
  const variants = await listActiveVariantsForProduct(product.id)

  const variantCards = (
    await Promise.all(
      variants.map(async (v) => {
        const skus = await listActiveSkusForVariant(v.id)
        if (!skus[0]) return null

        const skuCards = await Promise.all(
          skus.map(async (s) => {
            const [price, inStock] = await Promise.all([getSkuPrice(s.id), isSkuInStock(s.id)])
            return {
              sku: s.sku,
              attributes: filterPublic(s.attributes, publicKeys),
              price: price ? { amount: price.priceAmount, currency: price.currency } : null,
              inStock,
            }
          }),
        )

        return {
          imageIds: v.imageIds,
          attributes: filterPublic(v.attributes, publicKeys),
          skus: skuCards,
        }
      }),
    )
  ).filter(Boolean)
  if (!variantCards[0]) return null

  const representativeVariant = variantCards[0]
  const representativeSku = representativeVariant.skus[0]

  return {
    id: product.id,
    sku: representativeSku.sku,
    name: product.name,
    imageIds: representativeVariant.imageIds,
    price: representativeSku.price,
    inStock: representativeSku.inStock,
    variants: variantCards,
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
async function loadFilteredPage({ categoryId, cursor, limit, sortField, sortDirection, wantInStock, publicKeys }) {
  const items = []
  let nextCursor = cursor
  for (let i = 0; i < MAX_RAW_PAGES && items.length < limit; i++) {
    const response = await listActiveProducts({ categoryId, cursor: nextCursor, limit, sortField, sortDirection })
    const rawItems = response.items || []
    nextCursor = response.nextCursor || null

    const cards = (await Promise.all(rawItems.map((p) => toCard(p, publicKeys)))).filter(Boolean)
    for (const card of cards) {
      if (card.inStock === wantInStock) items.push(card)
    }

    if (!nextCursor) break
  }
  return { items, nextCursor }
}

// GET /api/catalog/products?categoryId=&cursor=&limit=&sort=&inStock= — TZ §8.2.
// Response is trimmed to card fields only, keyed off each product's
// representative variant/SKU with every sibling variant's SKUs and their
// own price/stock alongside (see toCard). inStock=true|false filters by
// the representative SKU's warehouse stock (see
// catalogClient.isSkuInStock); omitted means no stock filter. total is
// only reliable when unfiltered — a stock filter makes ProductsService's
// count meaningless, so it's dropped.
export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const sort = SORT_OPTIONS[query.sort?.toString()] || SORT_OPTIONS.in_stock
  const categoryId = query.categoryId?.toString()
  const cursor = query.cursor?.toString()
  const limit = query.limit ? Number(query.limit) : 50
  const inStockParam = query.inStock?.toString()

  const publicKeys = new Set((await listPublicAttributeDefinitions()).map((d) => d.key))

  if (inStockParam === 'true' || inStockParam === 'false') {
    const { items, nextCursor } = await loadFilteredPage({
      categoryId,
      cursor,
      limit,
      ...sort,
      wantInStock: inStockParam === 'true',
      publicKeys,
    })
    return { items, total: null, nextCursor }
  }

  const response = await listActiveProducts({ categoryId, cursor, limit, ...sort })
  const items = (await Promise.all((response.items || []).map((p) => toCard(p, publicKeys)))).filter(Boolean)
  return { items, total: response.total ?? null, nextCursor: response.nextCursor || null }
})
