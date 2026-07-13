// Basis points (backend/wire) <-> plain currency amount (what the user
// types), per TD §5.4 — never send/display raw basis points to the user.
export function toAmount(basisPoints) {
  return basisPoints == null ? null : basisPoints / 100
}

export function toBasisPoints(amount) {
  return amount == null ? null : Math.round(amount * 100)
}
