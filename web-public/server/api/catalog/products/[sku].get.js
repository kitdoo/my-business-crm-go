import {
  getActiveSkuBySku,
  getActiveVariantById,
  getActiveProductsByIds,
  listActiveVariantsForProduct,
  listActiveSkusForVariant,
  getSkuPrice,
  isSkuInStock,
  listPublicAttributeDefinitions,
} from '~~/server/utils/catalogClient.js'

function filterPublic(details, publicKeys) {
  return Object.fromEntries(Object.entries(details || {}).filter(([key]) => publicKeys.has(key)))
}

// GET /api/catalog/products/[sku] — product page (TZ §8.6), addressed by
// SKU sku (the purchasable unit, matching the cart's own sku keying — see
// useCart.js). Details is filtered down to only publicly-visible
// characteristic keys here — never trust the backend's raw map, private
// keys must not reach a site visitor (see listPublicAttributeDefinitions).
export default defineEventHandler(async (event) => {
  const sku = getRouterParam(event, 'sku')
  const currentSku = await getActiveSkuBySku(sku)
  if (!currentSku) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  const currentVariant = await getActiveVariantById(currentSku.variantId)
  if (!currentVariant) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  const [products, siblingVariants, publicDefinitions] = await Promise.all([
    getActiveProductsByIds([currentVariant.productId]),
    listActiveVariantsForProduct(currentVariant.productId),
    listPublicAttributeDefinitions(),
  ])
  const product = products[0]
  if (!product) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  const publicKeys = new Set(publicDefinitions.map((d) => d.key))

  // Every sibling variant's SKUs, each with its own price/stock, so the
  // page's two-tier picker (variant swatches, then SKU pills) can switch
  // price/availability/photo in place without a full navigation. A
  // variant with zero active SKUs has nothing purchasable to switch to —
  // dropped here the same way products.get.js's toCard already drops it
  // from the catalog grid, otherwise picking it leaves the page with no
  // active SKU to fall back on and the picker throws.
  const variants = (
    await Promise.all(
      siblingVariants.map(async (v) => {
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

  const currentSkuCard = variants
    .find((v) => v.skus.some((s) => s.sku === currentSku.sku))
    ?.skus.find((s) => s.sku === currentSku.sku)

  // Admin-set SKU characteristic (see productAttributeDefinitions seed
  // migration "generation") — the same color/size at a different
  // generation is a different SKU (see tmp/PHOMI MATERIJALI.xlsx), so this
  // lives on ProductSKU.attributes, never on the product or derived from a
  // category — the page reads it off the active SKU's own attributes (see
  // activeGeneration in [sku].vue), so it's dropped here only to avoid
  // also showing up as a plain row in the generic details table below.
  const { generation: _generation, ...currentSkuDetails } = filterPublic(currentSku.attributes, publicKeys)

  const details = {
    ...filterPublic(product.details, publicKeys),
    ...filterPublic(currentVariant.attributes, publicKeys),
    ...currentSkuDetails,
  }

  return {
    sku: currentSku.sku,
    name: product.name,
    description: product.description,
    details,
    imageIds: currentVariant.imageIds,
    price: currentSkuCard?.price ?? null,
    inStock: currentSkuCard?.inStock ?? false,
    // Every variant (visual option) and, within each, every SKU (size/
    // thickness/... option) — the current one included, so the page can
    // highlight it among its options.
    variants,
  }
})
