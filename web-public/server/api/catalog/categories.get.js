import { listActiveCategories } from '~~/server/utils/catalogClient.js'

// GET /api/catalog/categories — chips/tabs above the catalog grid (TZ §8.5).
export default defineEventHandler(async () => {
  const items = await listActiveCategories()
  // description is CMS-editable copy for the category's own catalog page
  // (see katalog/kategorija/[category].vue) — optional, falls back to the
  // generic catalog.intro on the frontend when empty.
  return { items: items.map((c) => ({ id: c.id, name: c.name, description: c.description })) }
})
