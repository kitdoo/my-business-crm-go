// Copied from web/app/composables/useFormat.js (admin) — only the money
// formatter is needed here (TZ §9), catalog prices come as basis points +
// ISO currency code same as ProductPrice on the admin side.
export function useFormatMoney() {
  const { locale } = useI18n()

  /** @param {number} basisPoints @param {string} currencyCode */
  function formatMoney(basisPoints, currencyCode) {
    if (basisPoints == null || !currencyCode) return ''
    const amount = basisPoints / 100
    return new Intl.NumberFormat(locale.value, { style: 'currency', currency: currencyCode }).format(amount)
  }

  return { formatMoney }
}
