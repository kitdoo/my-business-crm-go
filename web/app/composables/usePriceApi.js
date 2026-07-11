// ProductPrice isn't a generic EntityRegistry entity (TD §9.2/§12.1) — Get
// is keyed by productId, not id, and it lives as a tab on the Product
// detail page rather than its own list/route. Talks only to the
// server/api/prices/*.js proxy routes, same as useEntityApi does for
// generic entities.
export function usePriceApi() {
  return {
    async get(productId) {
      return $fetch('/api/prices/get', { method: 'POST', body: { productId } })
    },
    async create(productId, priceAmount, discountAmount) {
      return $fetch('/api/prices/create', { method: 'POST', body: { productId, priceAmount, discountAmount } })
    },
    async update(id, fields, mask, etag) {
      return $fetch('/api/prices/update', { method: 'PATCH', body: { id, fields, mask, etag } })
    },
    async history(productId, params = {}) {
      return $fetch('/api/prices/history', { method: 'POST', body: { productId, ...params } })
    },
  }
}
