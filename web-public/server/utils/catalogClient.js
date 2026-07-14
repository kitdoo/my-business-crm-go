import { getServiceClient, grpcCall } from './grpcClient.js'

// Server-only catalog reads (TZ §2/§8.2). ProductsService.List,
// ProductVariantsService.List, PricesService.Get and
// CategoriesService.List are anonymous on the backend (see
// internal/transports/grpc/interceptors/auth.New) — public site visitors
// are genuinely unauthenticated, there is no service/guest account
// involved. Safe only because the gRPC port is never reachable from
// outside the two Nitro BFFs (web/, web-public/); nothing on the backend
// restricts what filter an anonymous caller passes, so this file is the
// one place responsible for always forcing statuses:[ACTIVE].
//
// Product is a catalog card grouping one or more ProductVariant — the
// purchasable unit with its own sku, images, price (ProductPrice) and
// stock (Inventory), all keyed by variantId. The catalog listing shows a
// "representative" variant per product (first active one, by creation
// order); the product page (/katalog/:sku) resolves by variant sku and
// lets the visitor switch between sibling variants.

const PRODUCT_ACTIVE_STATUS = 'PRODUCT_STATUS_ACTIVE'
const VARIANT_ACTIVE_STATUS = 'PRODUCT_VARIANT_STATUS_ACTIVE'

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
    pagination: { limit: opts.limit || 24, cursor: opts.cursor },
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

/** @param {string} sku */
export async function getActiveVariantBySku(sku) {
  const response = await call('product_variant.proto', 'crm.grpc.product_variant.v1.ProductVariantsService', 'List', {
    filter: { statuses: [VARIANT_ACTIVE_STATUS], skus: [sku] },
    pagination: { limit: 1 },
  })
  return response.items?.[0] || null
}

/** @param {string} variantId */
export async function getVariantPrice(variantId) {
  try {
    const response = await call('price.proto', 'crm.grpc.price.v1.PricesService', 'Get', { variantId })
    return response.price || null
  } catch (err) {
    if (err?.code === 5 /* NOT_FOUND */) return null
    throw err
  }
}

// Sums Inventory.quantity across every warehouse for a variant and reduces
// it to a boolean — the exact per-warehouse quantity must never reach a
// site visitor, see the exemption comment on InventoryService.List in
// internal/transports/grpc/interceptors/auth.New.
export async function isVariantInStock(variantId) {
  const response = await call('inventory.proto', 'crm.grpc.inventory.v1.InventoryService', 'List', {
    variantId,
    pagination: { limit: 200 },
  })
  return (response.items || []).some((inv) => Number(inv.quantity) > 0)
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
// keys out of Product.details/ProductVariant.attributes before it reaches
// a site visitor.
export async function listPublicAttributeDefinitions() {
  const response = await call(
    'product_attribute_definition.proto',
    'crm.grpc.product_attribute_definition.v1.ProductAttributeDefinitionsService',
    'List',
    { filter: { isPublic: true }, pagination: { limit: 200 } },
  )
  return response.items || []
}
