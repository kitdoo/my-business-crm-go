import { getServiceClient, grpcCall } from './grpcClient.js'

// Server-only catalog reads (TZ §2/§8.2). ProductsService.List,
// ProductVariantsService.List, ProductSKUsService.List, PricesService.Get
// and CategoriesService.List are anonymous on the backend (see
// internal/transports/grpc/interceptors/auth.New) — public site visitors
// are genuinely unauthenticated, there is no service/guest account
// involved. Safe only because the gRPC port is never reachable from
// outside the two Nitro BFFs (web/, web-public/); nothing on the backend
// restricts what filter an anonymous caller passes, so this file is the
// one place responsible for always forcing statuses:[ACTIVE].
//
// Product is a catalog card grouping one or more ProductVariant — the
// visual identity (color/texture/pattern, images), no price or stock of
// its own. Each Variant groups one or more ProductSKU — the purchasable
// unit with its own sku, price (ProductPrice) and stock (Inventory), all
// keyed by skuId. The catalog listing shows a "representative" variant
// per product and its "representative" SKU (first active one of each, by
// creation order); the product page (/katalog/:sku) resolves by SKU sku
// and lets the visitor switch between sibling variants and sibling SKUs.

const PRODUCT_ACTIVE_STATUS = 'PRODUCT_STATUS_ACTIVE'
const VARIANT_ACTIVE_STATUS = 'PRODUCT_VARIANT_STATUS_ACTIVE'
const SKU_ACTIVE_STATUS = 'PRODUCT_SKU_STATUS_ACTIVE'

async function call(protoFile, servicePath, methodName, request) {
  const client = getServiceClient(protoFile, servicePath)
  return grpcCall(client, methodName, request)
}

/** @param {{ categoryId?: string, cursor?: string, limit?: number, sortField?: string, sortDirection?: string }} opts */
export async function listActiveProducts(opts = {}) {
  const request = {
    filter: {
      statuses: [PRODUCT_ACTIVE_STATUS], // always forced here, never trust the caller — TZ §1.1
      categoryIds: opts.categoryId ? [opts.categoryId] : [],
    },
    pagination: { limit: opts.limit || 50, cursor: opts.cursor },
    sort: {
      field: opts.sortField || 'FIELD_CREATED_AT',
      direction: opts.sortDirection || 'SORT_DIRECTION_DESC',
    },
    options: { includeTotalCount: true },
  }
  return call('product.proto', 'crm.grpc.product.v1.ProductsService', 'List', request)
}

/** @param {string[]} ids */
export async function getActiveProductsByIds(ids) {
  if (!ids.length) return []
  const response = await call('product.proto', 'crm.grpc.product.v1.ProductsService', 'List', {
    filter: { statuses: [PRODUCT_ACTIVE_STATUS], ids },
    pagination: { limit: ids.length },
  })
  return response.items || []
}

/** @param {string} productId */
export async function listActiveVariantsForProduct(productId) {
  const response = await call('product_variant.proto', 'crm.grpc.product_variant.v1.ProductVariantsService', 'List', {
    filter: { statuses: [VARIANT_ACTIVE_STATUS], productIds: [productId] },
    pagination: { limit: 100 },
  })
  return response.items || []
}

// Batched form of listActiveVariantsForProduct — one round trip for every
// product on a catalog page instead of one per product (ProductVariantsService.List
// already supports filter.productIds as an array). limit is generous
// because it bounds the whole page's variants, not one product's.
/** @param {string[]} productIds */
export async function listActiveVariantsForProducts(productIds) {
  if (!productIds.length) return []
  const response = await call('product_variant.proto', 'crm.grpc.product_variant.v1.ProductVariantsService', 'List', {
    filter: { statuses: [VARIANT_ACTIVE_STATUS], productIds },
    pagination: { limit: 500 },
  })
  return response.items || []
}

// Resolves a variant by id (no `ids` filter exists on the List RPC, so this
// is the only way to look one up directly) — used to walk a ProductSKU's
// variantId back to its owning Variant on the product page. Unlike List,
// Get applies no status filter of its own, so the ACTIVE check happens
// here — never trust its result otherwise.
/** @param {string} id */
export async function getActiveVariantById(id) {
  try {
    const response = await call('product_variant.proto', 'crm.grpc.product_variant.v1.ProductVariantsService', 'Get', { id })
    const variant = response.variant
    return variant && variant.status === VARIANT_ACTIVE_STATUS ? variant : null
  } catch (err) {
    if (err?.code === 5 /* NOT_FOUND */) return null
    throw err
  }
}

/** @param {string} variantId */
export async function listActiveSkusForVariant(variantId) {
  const response = await call('product_sku.proto', 'crm.grpc.product_sku.v1.ProductSKUsService', 'List', {
    filter: { statuses: [SKU_ACTIVE_STATUS], variantIds: [variantId] },
    pagination: { limit: 100 },
  })
  return response.items || []
}

// Batched form of listActiveSkusForVariant — one round trip for every
// variant on a catalog page instead of one per variant (ProductSKUsService.List
// already supports filter.variantIds as an array).
/** @param {string[]} variantIds */
export async function listActiveSkusForVariants(variantIds) {
  if (!variantIds.length) return []
  const response = await call('product_sku.proto', 'crm.grpc.product_sku.v1.ProductSKUsService', 'List', {
    filter: { statuses: [SKU_ACTIVE_STATUS], variantIds },
    pagination: { limit: 500 },
  })
  return response.items || []
}

/** @param {string} sku */
export async function getActiveSkuBySku(sku) {
  const response = await call('product_sku.proto', 'crm.grpc.product_sku.v1.ProductSKUsService', 'List', {
    filter: { statuses: [SKU_ACTIVE_STATUS], skus: [sku] },
    pagination: { limit: 1 },
  })
  return response.items?.[0] || null
}

/** @param {string} skuId */
export async function getSkuPrice(skuId) {
  try {
    const response = await call('price.proto', 'crm.grpc.price.v1.PricesService', 'Get', { skuId })
    return response.price || null
  } catch (err) {
    if (err?.code === 5 /* NOT_FOUND */) return null
    throw err
  }
}

// Batched form of getSkuPrice — one round trip for every SKU on a catalog
// page instead of one per SKU. Returns a Map keyed by skuId; a SKU with no
// current price is simply absent (see PricesService.List).
/** @param {string[]} skuIds */
export async function listPricesBySkuIds(skuIds) {
  if (!skuIds.length) return new Map()
  const response = await call('price.proto', 'crm.grpc.price.v1.PricesService', 'List', {
    filter: { skuIds },
  })
  return new Map((response.items || []).map((p) => [p.skuId, p]))
}

// Sums Inventory.quantity across every warehouse for a SKU and reduces it
// to a boolean — the exact per-warehouse quantity must never reach a site
// visitor, see the exemption comment on InventoryService.List in
// internal/transports/grpc/interceptors/auth.New.
export async function isSkuInStock(skuId) {
  const response = await call('inventory.proto', 'crm.grpc.inventory.v1.InventoryService', 'List', {
    skuId,
    pagination: { limit: 200 },
  })
  return (response.items || []).some((inv) => Number(inv.quantity) > 0)
}

// Batched form of isSkuInStock — one round trip for every SKU on a catalog
// page instead of one per SKU. Returns a Set of skuIds that have stock in
// at least one warehouse.
/** @param {string[]} skuIds */
export async function listInStockSkuIds(skuIds) {
  if (!skuIds.length) return new Set()
  const response = await call('inventory.proto', 'crm.grpc.inventory.v1.InventoryService', 'List', {
    filter: { skuIds },
    pagination: { limit: 1000 },
  })
  const inStock = new Set()
  for (const inv of response.items || []) {
    if (Number(inv.quantity) > 0) inStock.add(inv.skuId)
  }
  return inStock
}

export async function listActiveCategories() {
  const response = await call('category.proto', 'crm.grpc.category.v1.CategoriesService', 'List', {
    filter: { statuses: ['CATEGORY_STATUS_ACTIVE'] },
    pagination: { limit: 100 },
  })
  return response.items || []
}

// Characteristics catalog, filtered to isPublic: true here — never trust
// the caller, same as PRODUCT_ACTIVE_STATUS above. Used to strip private
// keys out of Product.details/ProductVariant.attributes/ProductSKU.attributes
// before it reaches a site visitor.
export async function listPublicAttributeDefinitions() {
  const response = await call(
    'product_attribute_definition.proto',
    'crm.grpc.product_attribute_definition.v1.ProductAttributeDefinitionsService',
    'List',
    { filter: { isPublic: true }, pagination: { limit: 200 } },
  )
  return response.items || []
}
