import {
  getActiveVariantBySku,
  getActiveProductsByIds,
  listActiveVariantsForProduct,
  getVariantPrice,
  listPublicAttributeDefinitions,
} from '~~/server/utils/catalogClient.js'

function filterPublic(details, publicKeys) {
  return Object.fromEntries(Object.entries(details || {}).filter(([key]) => publicKeys.has(key)))
}

// GET /api/catalog/products/[sku] — product page (TZ §8.6), addressed by
// variant sku (the purchasable unit, matching the cart's own sku keying —
// see useCart.js). Details is filtered down to only publicly-visible
// characteristic keys here — never trust the backend's raw map, private
// keys must not reach a site visitor (see listPublicAttributeDefinitions).
export default defineEventHandler(async (event) => {
  const sku = getRouterParam(event, 'sku')
  const variant = await getActiveVariantBySku(sku)
  if (!variant) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  const [products, siblingVariants, price, publicDefinitions] = await Promise.all([
    getActiveProductsByIds([variant.productId]),
    listActiveVariantsForProduct(variant.productId),
    getVariantPrice(variant.id),
    listPublicAttributeDefinitions(),
  ])
  const product = products[0]
  if (!product) {
    throw createError({ statusCode: 404, statusMessage: 'Not found' })
  }

  const publicKeys = new Set(publicDefinitions.map((d) => d.key))
  const details = {
    ...filterPublic(product.details, publicKeys),
    ...filterPublic(variant.attributes, publicKeys),
  }

  return {
    id: variant.id,
    productId: product.id,
    sku: variant.sku,
    name: product.name,
    description: product.description,
    details,
    imageIds: variant.imageIds,
    price: price ? { amount: price.priceAmount, currency: price.currency } : null,
    status: variant.status,
    // Sibling variants for the color/size switcher — the current one
    // included, so the page can highlight it among its options.
    variants: siblingVariants.map((v) => ({
      sku: v.sku,
      imageIds: v.imageIds,
      attributes: filterPublic(v.attributes, publicKeys),
    })),
  }
})
