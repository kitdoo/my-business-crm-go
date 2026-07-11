import { defineStore } from 'pinia'
import { getEntityConfig } from '~/config/entities'

// Resolves <entityKey, id> -> display label, batching same-tick requests
// into one `list` call where the entity supports an ids filter (TD §9.5).
// Session-lifetime cache only — not persisted.
export const useReferenceCacheStore = defineStore('referenceCache', {
  state: () => ({
    cache: {}, // { [entityKey]: { [id]: item } }
    pendingBatches: {}, // { [entityKey]: { ids: Set, timer, resolvers: Map<id, fn[]> } }
  }),
  actions: {
    getCached(entityKey, id) {
      return this.cache[entityKey]?.[id]
    },

    setCached(entityKey, item) {
      if (!this.cache[entityKey]) this.cache[entityKey] = {}
      this.cache[entityKey][item.id] = item
    },

    async resolve(entityKey, id) {
      if (!id) return null
      const cached = this.getCached(entityKey, id)
      if (cached) return cached

      if (!this.pendingBatches[entityKey]) {
        this.pendingBatches[entityKey] = { ids: new Set(), resolvers: new Map(), timer: null }
      }
      const batch = this.pendingBatches[entityKey]
      batch.ids.add(id)

      const promise = new Promise((resolve) => {
        const list = batch.resolvers.get(id) || []
        list.push(resolve)
        batch.resolvers.set(id, list)
      })

      if (!batch.timer) {
        batch.timer = setTimeout(() => this._flushBatch(entityKey), 0)
      }
      return promise
    },

    async _flushBatch(entityKey) {
      const batch = this.pendingBatches[entityKey]
      delete this.pendingBatches[entityKey]
      const ids = Array.from(batch.ids)
      const config = getEntityConfig(entityKey)
      const filterKey = config?.list?.idsFilterKey

      try {
        if (filterKey) {
          const { list } = useEntityApi(entityKey)
          const { items } = await list({ filter: { [filterKey]: ids }, pagination: { limit: ids.length } })
          for (const item of items) this.setCached(entityKey, item)
        } else {
          const { get } = useEntityApi(entityKey)
          const items = await Promise.all(ids.map((id) => get(id)))
          for (const item of items) if (item) this.setCached(entityKey, item)
        }
      } finally {
        for (const id of ids) {
          const item = this.getCached(entityKey, id) || null
          const resolvers = batch.resolvers.get(id) || []
          for (const resolve of resolvers) resolve(item)
        }
      }
    },
  },
})
