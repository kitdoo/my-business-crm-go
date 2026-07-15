// ProductPrice isn't a generic EntityRegistry entity (TD §9.2/§12.1) — Get
// is keyed by skuId, not id, and it lives inside the SKU's own detail
// page (Price tab) rather than its own list/route. Talks only to the
// server/api/prices/*.js proxy routes, same as useEntityApi does for
// generic entities.
export function usePriceApi() {
  return {
    async get(skuId) {
      return $fetch('/api/prices/get', { method: 'POST', body: { skuId } })
    },
    async create(skuId, priceAmount, discountAmount) {
      return $fetch('/api/prices/create', { method: 'POST', body: { skuId, priceAmount, discountAmount } })
    },
    async update(id, fields, mask, etag) {
      return $fetch('/api/prices/update', { method: 'PATCH', body: { id, fields, mask, etag } })
    },
    async history(skuId, params = {}) {
      return $fetch('/api/prices/history', { method: 'POST', body: { skuId, ...params } })
    },
  }
}
