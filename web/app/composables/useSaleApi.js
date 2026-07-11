// Sale isn't a generic EntityRegistry entity (TD §12.3) — items are
// immutable, status moves only through UpdateStatus/Cancel, no generic
// Update/Delete. `list` still goes through the standard useEntityApi
// naming convention (POST /api/sales/list), so <EntityDataTable
// entity="sales"> works unmodified for the list page; this composable
// covers the rest.
export function useSaleApi() {
  return {
    async get(id) {
      return $fetch('/api/sales/get', { method: 'POST', body: { id } })
    },
    async create(fields) {
      return $fetch('/api/sales/create', { method: 'POST', body: fields })
    },
    async updateStatus(id, status, etag) {
      return $fetch('/api/sales/update-status', { method: 'POST', body: { id, status, etag } })
    },
    async cancel(id, reason, etag) {
      return $fetch('/api/sales/cancel', { method: 'POST', body: { id, reason, etag } })
    },
  }
}
