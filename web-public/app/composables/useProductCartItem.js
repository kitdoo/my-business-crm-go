import { localizedText } from '~/utils/localizedText.js'

// Wires the active variant+SKU selection to the cart — shared by ProductCard
// and the product detail page, which both need the same "combine
// variant+SKU attributes into one label" and qty-stepper behavior.
// `product`, `activeVariant`, `activeSku` are refs/computeds the caller owns.
export function useProductCartItem(product, activeVariant, activeSku) {
  const { locale } = useI18n()
  const { addItem, setQty, getQty, removeItem } = useCart()

  function cartOptions() {
    const attrs = { ...(activeVariant.value?.attributes || {}), ...(activeSku.value?.attributes || {}) }
    return Object.values(attrs).map((v) => localizedText(v, locale.value)).filter(Boolean).join(' / ')
  }

  function cartItem() {
    return { ...product.value, sku: activeSku.value.sku, price: activeSku.value.price, options: cartOptions() }
  }

  const qtyInCart = computed(() => getQty(activeSku.value.sku))

  function increment() {
    if (qtyInCart.value) setQty(activeSku.value.sku, qtyInCart.value + 1)
    else addItem(cartItem())
  }

  function decrement() {
    if (qtyInCart.value <= 1) removeItem(activeSku.value.sku)
    else setQty(activeSku.value.sku, qtyInCart.value - 1)
  }

  return { qtyInCart, cartItem, addItem, increment, decrement }
}
