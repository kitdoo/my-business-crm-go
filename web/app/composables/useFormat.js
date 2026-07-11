// Three shared formatters (TD §5.4/§9.6) — no inline `new Date(...)` or
// `toFixed(2)` in components, everything goes through these.

export function useFormatDate() {
  const { locale } = useI18n()

  /**
   * @param {number | null | undefined} unixSeconds
   * @param {'short' | 'long' | 'relative'} [style]
   */
  function formatDate(unixSeconds, style = 'short') {
    if (!unixSeconds) return ''
    const date = new Date(unixSeconds * 1000)
    if (style === 'relative') {
      const rtf = new Intl.RelativeTimeFormat(locale.value, { numeric: 'auto' })
      const diffSeconds = Math.round((date.getTime() - Date.now()) / 1000)
      const divisions = [
        { amount: 60, unit: 'second' },
        { amount: 60, unit: 'minute' },
        { amount: 24, unit: 'hour' },
        { amount: 30, unit: 'day' },
        { amount: 12, unit: 'month' },
        { amount: Infinity, unit: 'year' },
      ]
      let duration = diffSeconds
      for (const division of divisions) {
        if (Math.abs(duration) < division.amount) return rtf.format(Math.round(duration), division.unit)
        duration /= division.amount
      }
    }
    const options = style === 'long'
      ? { dateStyle: 'long', timeStyle: 'short' }
      : { dateStyle: 'short' }
    return new Intl.DateTimeFormat(locale.value, options).format(date)
  }

  return { formatDate }
}

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

export function useFormatPercent() {
  function formatPercent(value) {
    if (value == null) return ''
    return `${value}%`
  }
  return { formatPercent }
}
