<script setup>
// Comma-joined warehouse names used by a sale's line items (TD §12.3) —
// each SaleItem now carries its own warehouseId (a sale can draw
// different lines from different warehouses), so there's no single
// top-level warehouseId to resolve via RelationLabel/RelationListLabel
// anymore; this dedupes across value (the sale's items array) instead.
const props = defineProps({
  value: { type: Array, default: () => [] }, // SaleItem[]
})

const warehouseIds = computed(() => [...new Set((props.value || []).map((item) => item.warehouseId).filter(Boolean))])
</script>

<template>
  <span v-if="!warehouseIds.length" />
  <span v-else>
    <template v-for="(id, index) in warehouseIds" :key="id">
      <RelationLabel :value="id" relation="warehouses" />{{ index < warehouseIds.length - 1 ? ', ' : '' }}
    </template>
  </span>
</template>
