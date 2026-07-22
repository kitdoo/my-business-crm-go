// Total stock for one SKU, summed across every warehouse (TD §12.4) — used
// wherever a SKU needs to show "how much is there" as a plain number
// instead of a per-warehouse breakdown (ProductSkuGeneralForm.vue).
export async function fetchTotalStock(inventoryApi, skuId) {
  const res = await inventoryApi.list({ filter: { skuId }, pagination: { limit: 200 } })
  return (res.items || []).reduce((sum, item) => sum + (item.quantity || 0), 0)
}

// Total stock for several SKUs in one round trip, via filter.skuIds
// (batch lookup — see InventoryListRequest.Filter) instead of one
// inventoryApi.list call per SKU. Used by ProductSkusPanel.vue and
// ProductVariantsReadOnly.vue, which each show stock for many SKUs at
// once. Returns { [skuId]: totalQuantity }, defaulting missing/zero-stock
// SKUs to 0.
export async function fetchTotalStockBySkuIds(inventoryApi, skuIds) {
  const totals = Object.fromEntries(skuIds.map((id) => [id, 0]))
  if (skuIds.length === 0) return totals
  const res = await inventoryApi.list({ filter: { skuIds }, pagination: { limit: 1000 } })
  for (const item of res.items || []) {
    totals[item.skuId] = (totals[item.skuId] || 0) + (item.quantity || 0)
  }
  return totals
}
