// Shared between the sale detail page and the sales list (row action
// icons) — a sale in one of these statuses accepts no further status
// change or cancellation; the server is the authoritative check, this is
// UX only.
export const SALE_TERMINAL_STATUSES = ['SALE_STATUS_CANCELLED', 'SALE_STATUS_REFUNDED']

export function isSaleTerminal(status) {
  return SALE_TERMINAL_STATUSES.includes(status)
}
