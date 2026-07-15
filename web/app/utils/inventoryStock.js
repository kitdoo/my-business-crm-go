// Total stock for one SKU, summed across every warehouse (TD §12.4) — used
// wherever a SKU needs to show "how much is there" as a plain number
// instead of a per-warehouse breakdown (ProductSkuGeneralForm.vue,
// ProductSkusPanel.vue, ProductVariantsReadOnly.vue). Inventory.list only
// filters by a single skuId (see entities/inventory.go's InventoryList —
// no skuIds plural), so callers wanting several SKUs' totals still need
// one call per SKU.
export async function fetchTotalStock(inventoryApi, skuId) {
  const res = await inventoryApi.list({ filter: { skuId }, pagination: { limit: 200 } })
  return (res.items || []).reduce((sum, item) => sum + (item.quantity || 0), 0)
}
