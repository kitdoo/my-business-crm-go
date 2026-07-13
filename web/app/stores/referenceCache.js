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
        const { list, get } = useEntityApi(entityKey)
        let missing = ids
        if (filterKey) {
          const { items } = await list({ filter: { [filterKey]: ids }, pagination: { limit: ids.length } })
          for (const item of items) this.setCached(entityKey, item)
          missing = ids.filter((id) => !this.getCached(entityKey, id))
        }
        // Falls back to per-id Get for anything the batch didn't resolve —
        // covers both the no-filterKey case and a filterKey the backend
        // doesn't actually honor (silently returning an unrelated page
        // instead of filtering), so a config mistake degrades to slightly
        // more requests instead of blank labels. allSettled, not all — one
        // stale/deleted id 404ing must not blank out every other label in
        // the same batch tick.
        if (missing.length > 0) {
          const results = await Promise.allSettled(missing.map((id) => get(id)))
          for (const result of results) {
            if (result.status === 'fulfilled' && result.value) this.setCached(entityKey, result.value)
          }
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
