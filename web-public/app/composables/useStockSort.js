// Shared "in-stock options first" ordering — used for both variant swatches
// and SKU pills, on the catalog grid (ProductCard) and the product detail
// page ([sku].vue) alike. `inStock === null` (not yet loaded, see
// CatalogPage's staged fetch) sorts as "not in stock" so the list settles
// into its final order once, rather than jumping mid-load.
export function useStockSort() {
  function stockScore(inStock) {
    return inStock === true ? 0 : 1
  }

  // Array.prototype.sort is stable (ES2019+), so items tied on stockScore
  // keep their original relative order.
  function sortIndicesByStock(list, getInStock) {
    return list.map((_, i) => i).sort((a, b) => stockScore(getInStock(list[a])) - stockScore(getInStock(list[b])))
  }

  function firstInStockIndex(list, getInStock) {
    const idx = list.findIndex((item) => getInStock(item) === true)
    return idx === -1 ? 0 : idx
  }

  return { sortIndicesByStock, firstInStockIndex }
}
