// Hundredths of a unit (backend/wire) <-> plain quantity (what the user
// types) — same basis-points convention as priceAmount.js — never send/
// display raw hundredths to the user.
export function toQuantity(basisQuantity) {
  return basisQuantity == null ? null : basisQuantity / 100
}

export function toBasisQuantity(quantity) {
  return quantity == null ? null : Math.round(quantity * 100)
}
