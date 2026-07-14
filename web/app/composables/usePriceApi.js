// ProductPrice isn't a generic EntityRegistry entity (TD §9.2/§12.1) — Get
// is keyed by variantId, not id, and it lives inside the variant's edit
// drawer on the Product detail page's Variants tab rather than its own
// list/route. Talks only to the server/api/prices/*.js proxy routes, same
// as useEntityApi does for generic entities.
export function usePriceApi() {
  return {
    async get(variantId) {
      return $fetch('/api/prices/get', { method: 'POST', body: { variantId } })
    },
    async create(variantId, priceAmount, discountAmount) {
      return $fetch('/api/prices/create', { method: 'POST', body: { variantId, priceAmount, discountAmount } })
    },
    async update(id, fields, mask, etag) {
      return $fetch('/api/prices/update', { method: 'PATCH', body: { id, fields, mask, etag } })
    },
    async history(variantId, params = {}) {
      return $fetch('/api/prices/history', { method: 'POST', body: { variantId, ...params } })
    },
  }
}
