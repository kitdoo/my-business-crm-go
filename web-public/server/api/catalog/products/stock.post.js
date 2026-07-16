import { loadPriceAndStockBySkuId } from '~~/server/api/catalog/products.get.js'

// POST /api/catalog/products/stock, body: { skuIds: string[] } — the second
// stage of the catalog page's staged load (see katalog/index.vue and
// products.get.js's includePriceStock). Called once the card grid has
// already rendered from the first, price/stock-free response, with the
// skuIds of every SKU currently on screen (products.get.js stamps a skuId
// onto each sku card for exactly this).
export default defineEventHandler(async (event) => {
  const body = await readBody(event)
  const skuIds = Array.isArray(body?.skuIds) ? [...new Set(body.skuIds.filter((id) => typeof id === 'string' && id))] : []
  if (!skuIds.length) return { items: [] }

  const bySkuId = await loadPriceAndStockBySkuId(skuIds)
  return {
    items: skuIds.map((skuId) => ({ skuId, ...(bySkuId.get(skuId) || { price: null, inStock: false }) })),
  }
})
