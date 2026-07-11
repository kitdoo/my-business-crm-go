export function useCatalogApi() {
  function listProducts(params = {}) {
    return $fetch('/api/catalog/products', { query: params })
  }
  function getProduct(sku) {
    return $fetch(`/api/catalog/products/${sku}`)
  }
  function listCategories() {
    return $fetch('/api/catalog/categories')
  }
  return { listProducts, getProduct, listCategories }
}
