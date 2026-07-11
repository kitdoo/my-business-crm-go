// Reports isn't a generic EntityRegistry entity — six independent
// read-only methods, one per dashboard widget (TD §8.4), each of which
// loads and fails on its own so one broken report doesn't blank the
// whole dashboard.
export function useReportApi() {
  return {
    async salesReport(period) {
      return $fetch('/api/reports/sales-report', { method: 'POST', body: { period } })
    },
    async salesByStaff(period) {
      return $fetch('/api/reports/sales-by-staff', { method: 'POST', body: { period } })
    },
    async salesByPartner(period) {
      return $fetch('/api/reports/sales-by-partner', { method: 'POST', body: { period } })
    },
    async popularProducts(period, limit = 10) {
      return $fetch('/api/reports/popular-products', { method: 'POST', body: { period, limit } })
    },
    async turnover(period) {
      return $fetch('/api/reports/turnover', { method: 'POST', body: { period } })
    },
    async stockLevels(warehouseId) {
      return $fetch('/api/reports/stock-levels', { method: 'POST', body: { warehouseId } })
    },
  }
}
