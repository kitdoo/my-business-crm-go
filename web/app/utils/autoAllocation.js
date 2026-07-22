// Greedily splits a requested quantity across warehouse stock, best-stocked
// warehouse first — minimizes the number of allocation rows produced for
// the common case. Pure and synchronous: pages/sales/new.vue fetches the
// stock list (Inventory.List) and passes it in here, so the algorithm
// itself stays unit-testable without mocking the network.
//
// `stock` is a list of { warehouseId, quantity }; quantity may arrive as a
// numeric string (grpc-js's longs:'String' mode), so it's coerced with
// Number() rather than compared/subtracted directly.
//
// Returns { allocations, shortfall }: shortfall is 0 on success, or the
// remaining unallocated quantity when stock runs out across every
// warehouse (allocations is null in that case).
export function resolveAutoAllocation(quantity, stock) {
  const sorted = (stock || [])
    .map((s) => ({ warehouseId: s.warehouseId, quantity: Number(s.quantity) }))
    .filter((s) => s.quantity > 0)
    .sort((a, b) => b.quantity - a.quantity)

  let remaining = Number(quantity)
  const allocations = []
  for (const s of sorted) {
    if (remaining <= 0) break
    const take = Math.min(remaining, s.quantity)
    allocations.push({ warehouseId: s.warehouseId, quantity: take })
    remaining -= take
  }

  if (remaining > 0) {
    return { allocations: null, shortfall: remaining }
  }
  return { allocations, shortfall: 0 }
}
