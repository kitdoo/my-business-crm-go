export function useCatalogApi() {
  function listProducts(params = {}) {
    return $fetch('/api/catalog/products', { query: params })
  }
  function getProduct(sku) {
    return $fetch(`/api/catalog/products/${sku}`)
  }
  function getProductsStock(skuIds) {
    return $fetch('/api/catalog/products/stock', { method: 'POST', body: { skuIds } })
  }
  function listCategories() {
    return $fetch('/api/catalog/categories')
  }
  return { listProducts, getProduct, getProductsStock, listCategories }
}
