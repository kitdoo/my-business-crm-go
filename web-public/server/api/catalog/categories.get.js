import { listActiveCategories } from '~~/server/utils/catalogClient.js'

// GET /api/catalog/categories — chips/tabs above the catalog grid (TZ §8.5).
export default defineEventHandler(async () => {
  const items = await listActiveCategories()
  return { items: items.map((c) => ({ id: c.id, name: c.name })) }
})
