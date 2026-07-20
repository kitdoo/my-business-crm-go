import {
  listActiveProducts,
  listActiveVariantsForProducts,
  listActiveSkusForVariants,
  listPricesBySkuIds,
  listInStockSkuIds,
  listPublicAttributeDefinitions,
} from '~~/server/utils/catalogClient.js'

function filterPublic(details, publicKeys) {
  return Object.fromEntries(Object.entries(details || {}).filter(([key]) => publicKeys.has(key)))
}

function groupBy(items, key) {
  const map = new Map()
  for (const item of items) {
    const list = map.get(item[key])
    if (list) list.push(item)
    else map.set(item[key], [item])
  }
  return map
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

// Looks up price + stock for a batch of SKU ids in two round trips total
// (PricesService.List, InventoryService.List — see catalogClient), keyed
// by skuId. Shared by the eager path (loadFilteredPage, which needs stock
// to decide what belongs on a filtered page) and the standalone
// products/stock.post.js endpoint the client calls once cards have
// already rendered without price/stock (see toCards's includePriceStock).
export async function loadPriceAndStockBySkuId(skuIds) {
  const [prices, inStockSet] = await Promise.all([listPricesBySkuIds(skuIds), listInStockSkuIds(skuIds)])
  const map = new Map()
  for (const skuId of skuIds) {
    const price = prices.get(skuId)
    map.set(skuId, {
      price: price ? { amount: price.priceAmount, currency: price.currency } : null,
      inStock: inStockSet.has(skuId),
    })
  }
  return map
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
//
// Builds cards for a whole batch of products at once: every product's
// variants are fetched in a single call (ProductVariantsService.List
// filter.productIds), then every variant's SKUs in another single call
// (ProductSKUsService.List filter.variantIds) — replacing what used to be
// one round trip per product and one per variant.
//
// includePriceStock controls whether price/stock is resolved eagerly
// here (needed by loadFilteredPage, which filters on inStock) or left
// null for the client to fetch afterwards via products/stock.post.js once
// the cards themselves — image, name, variant/SKU pickers — have already
// painted (see katalog/index.vue). Every sku card always carries its own
// skuId so that follow-up request can ask for exactly the SKUs on screen
// without re-resolving anything.
async function toCards(products, publicKeys, { includePriceStock }) {
  if (!products.length) return []

  const variants = await listActiveVariantsForProducts(products.map((p) => p.id))
  const variantsByProduct = groupBy(variants, 'productId')

  const skus = await listActiveSkusForVariants(variants.map((v) => v.id))
  const skusByVariant = groupBy(skus, 'variantId')

  const priceAndStockBySkuId = includePriceStock
    ? await loadPriceAndStockBySkuId(skus.map((s) => s.id))
    : new Map()

  return products
    .map((product) => {
      const variantCards = (variantsByProduct.get(product.id) || [])
        .map((v) => {
          const variantSkus = skusByVariant.get(v.id) || []
          if (!variantSkus[0]) return null
          const skuCards = variantSkus.map((s) => {
            const stock = priceAndStockBySkuId.get(s.id)
            return {
              skuId: s.id,
              sku: s.sku,
              attributes: filterPublic(s.attributes, publicKeys),
              price: stock ? stock.price : null,
              inStock: stock ? stock.inStock : null, // null = not resolved yet, see includePriceStock
            }
          })
          return { imageIds: v.imageIds, attributes: filterPublic(v.attributes, publicKeys), skus: skuCards }
        })
        .filter(Boolean)
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
        // Every SKU carries its own "generation" characteristic (see
        // productAttributeDefinitions seed migration "generation") inside
        // its attributes — the same color/size at different generations
        // are different SKUs (see tmp/PHOMI MATERIJALI.xlsx), so this is
        // never a per-product value and never derived from a category.
        variants: variantCards,
      }
    })
    .filter(Boolean)
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
// call; a returned page can therefore run a little over `limit`. Stock is
// resolved eagerly here (includePriceStock: true) since which cards even
// belong on the page depends on it.
async function loadFilteredPage({ categoryId, cursor, limit, sortField, sortDirection, wantInStock, publicKeys }) {
  const items = []
  let nextCursor = cursor
  for (let i = 0; i < MAX_RAW_PAGES && items.length < limit; i++) {
    const response = await listActiveProducts({ categoryId, cursor: nextCursor, limit, sortField, sortDirection })
    const rawItems = response.items || []
    nextCursor = response.nextCursor || null

    const cards = await toCards(rawItems, publicKeys, { includePriceStock: true })
    for (const card of cards) {
      if (card.inStock === wantInStock) items.push(card)
    }

    if (!nextCursor) break
  }
  return { items, nextCursor }
}

// GET /api/catalog/products?categoryId=&cursor=&limit=&sort=&inStock= — TZ §8.2.
// Response is trimmed to card fields only, keyed off each product's
// representative variant/SKU with every sibling variant's SKUs alongside
// (see toCards). inStock=true|false filters by the representative SKU's
// warehouse stock (see catalogClient.isSkuInStock); omitted means no stock
// filter. total is only reliable when unfiltered — a stock filter makes
// ProductsService's count meaningless, so it's dropped.
//
// Unfiltered requests skip price/stock entirely (cards render immediately
// with image/name/variant pickers) — the client fetches price/stock in a
// second pass via POST /api/catalog/products/stock once these cards have
// painted. A stock filter forces the eager path since price/stock decides
// which cards even qualify for the page.
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
  const items = await toCards(response.items || [], publicKeys, { includePriceStock: false })
  return { items, total: response.total ?? null, nextCursor: response.nextCursor || null }
})
