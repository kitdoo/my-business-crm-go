import { localizedText } from '~/utils/localizedText.js'

/** Shared cart state (item 10) — a plain useState singleton is enough here,
 * no Pinia store exists in this app yet. Persistence to localStorage is
 * handled once, on the client, by plugins/cart.client.js; this composable
 * only mutates the in-memory list. */
export function useCart() {
  const items = useState('cart-items', () => [])
  const isOpen = useState('cart-open', () => false)
  const { locale } = useI18n()

  function addItem(product, qty = 1) {
    const existing = items.value.find((i) => i.sku === product.sku)
    if (existing) {
      existing.qty += qty
      return
    }
    items.value.push({
      sku: product.sku,
      name: localizedText(product.name, locale.value) || product.sku,
      options: product.options || '',
      price: product.price || null,
      qty,
    })
  }

  function removeItem(sku) {
    items.value = items.value.filter((i) => i.sku !== sku)
  }

  function setQty(sku, qty) {
    const item = items.value.find((i) => i.sku === sku)
    if (item) item.qty = Math.max(1, Math.floor(qty) || 1)
  }

  function getQty(sku) {
    return items.value.find((i) => i.sku === sku)?.qty || 0
  }

  function clear() {
    items.value = []
  }

  function open() {
    isOpen.value = true
  }

  function close() {
    isOpen.value = false
  }

  const count = computed(() => items.value.reduce((sum, i) => sum + i.qty, 0))

  // Total across items that have a resolved price — items without one
  // (price not set on the product) are excluded rather than treated as 0,
  // so the total isn't silently wrong.
  const totalAmount = computed(() => items.value.reduce((sum, i) => sum + (i.price ? i.price.amount * i.qty : 0), 0))
  const totalCurrency = computed(() => items.value.find((i) => i.price)?.price.currency || null)

  // Plain-text product list for the checkout email's message field —
  // "текстовое поле где перечислены товары" (item 10).
  const summaryText = computed(() => items.value.map((i) => `${i.sku} — ${i.name}${i.options ? ` (${i.options})` : ''} × ${i.qty}`).join('\n'))

  return {
    items,
    isOpen,
    addItem,
    removeItem,
    setQty,
    getQty,
    clear,
    open,
    close,
    count,
    totalAmount,
    totalCurrency,
    summaryText,
  }
}
