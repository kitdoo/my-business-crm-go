// Hydrates the shared cart state from localStorage once at app start and
// keeps it in sync on every change — this is what makes the cart survive a
// closed browser (item 10). Client-only (.client.js) since localStorage
// doesn't exist during SSR.
const STORAGE_KEY = 'phomi-cart-items'

export default defineNuxtPlugin(() => {
  const items = useState('cart-items', () => [])

  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (Array.isArray(parsed)) items.value = parsed
    }
  } catch {
    // Corrupt/blocked storage — start with an empty cart rather than throwing.
  }

  watch(
    items,
    (val) => {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(val))
      } catch {
        // Storage full/blocked — cart still works for the current session.
      }
    },
    { deep: true },
  )
})
