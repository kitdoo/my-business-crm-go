// One factory for every entity's CRUD calls (TD §9.2). Talks only to
// same-origin Nuxt server routes — never to gRPC directly. Errors are not
// caught here; callers (usually useEntityForm / EntityListPage) route them
// through useApiErrorHandler().
export function useEntityApi(entityKey) {
  const base = `/api/${entityKey}`

  return {
    async list(params = {}) {
      return $fetch(`${base}/list`, { method: 'POST', body: params })
    },
    async get(id) {
      return $fetch(`${base}/get`, { method: 'POST', body: { id } })
    },
    async create(fields) {
      return $fetch(`${base}/create`, { method: 'POST', body: { fields } })
    },
    async update(id, fields, mask, etag) {
      return $fetch(`${base}/update`, { method: 'PATCH', body: { id, fields, mask, etag } })
    },
    async remove(id, etag) {
      return $fetch(`${base}/delete`, { method: 'DELETE', body: { id, etag } })
    },
  }
}
