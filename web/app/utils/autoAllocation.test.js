import { describe, expect, it } from 'vitest'
import { resolveAutoAllocation } from './autoAllocation.js'

describe('resolveAutoAllocation', () => {
  it('takes everything from a single warehouse when it covers the quantity', () => {
    const { allocations, shortfall } = resolveAutoAllocation(5, [{ warehouseId: 'w1', quantity: 10 }])
    expect(shortfall).toBe(0)
    expect(allocations).toEqual([{ warehouseId: 'w1', quantity: 5 }])
  })

  it('fills from the best-stocked warehouse first', () => {
    const stock = [
      { warehouseId: 'w1', quantity: 3 },
      { warehouseId: 'w2', quantity: 10 },
      { warehouseId: 'w3', quantity: 5 },
    ]
    const { allocations, shortfall } = resolveAutoAllocation(12, stock)
    expect(shortfall).toBe(0)
    expect(allocations).toEqual([
      { warehouseId: 'w2', quantity: 10 },
      { warehouseId: 'w3', quantity: 2 },
    ])
  })

  it('splits across every warehouse needed and stops once covered', () => {
    const stock = [
      { warehouseId: 'w1', quantity: 4 },
      { warehouseId: 'w2', quantity: 4 },
      { warehouseId: 'w3', quantity: 4 },
    ]
    const { allocations, shortfall } = resolveAutoAllocation(9, stock)
    expect(shortfall).toBe(0)
    expect(allocations).toHaveLength(3)
    expect(allocations.reduce((sum, a) => sum + a.quantity, 0)).toBe(9)
  })

  it('reports the shortfall when total stock falls short', () => {
    const stock = [
      { warehouseId: 'w1', quantity: 2 },
      { warehouseId: 'w2', quantity: 3 },
    ]
    const { allocations, shortfall } = resolveAutoAllocation(10, stock)
    expect(allocations).toBeNull()
    expect(shortfall).toBe(5)
  })

  it('reports the full quantity as shortfall with no stock at all', () => {
    const { allocations, shortfall } = resolveAutoAllocation(3, [])
    expect(allocations).toBeNull()
    expect(shortfall).toBe(3)
  })

  it('ignores zero and negative stock rows', () => {
    const stock = [
      { warehouseId: 'w1', quantity: 0 },
      { warehouseId: 'w2', quantity: -5 },
      { warehouseId: 'w3', quantity: 6 },
    ]
    const { allocations, shortfall } = resolveAutoAllocation(6, stock)
    expect(shortfall).toBe(0)
    expect(allocations).toEqual([{ warehouseId: 'w3', quantity: 6 }])
  })

  it('coerces stock quantities that arrive as numeric strings (grpc longs:String)', () => {
    const stock = [
      { warehouseId: 'w1', quantity: '4' },
      { warehouseId: 'w2', quantity: '10' },
    ]
    const { allocations, shortfall } = resolveAutoAllocation(12, stock)
    expect(shortfall).toBe(0)
    expect(allocations).toEqual([
      { warehouseId: 'w2', quantity: 10 },
      { warehouseId: 'w1', quantity: 2 },
    ])
  })

  it('requesting exactly zero needs no allocation', () => {
    const { allocations, shortfall } = resolveAutoAllocation(0, [{ warehouseId: 'w1', quantity: 5 }])
    expect(shortfall).toBe(0)
    expect(allocations).toEqual([])
  })
})
