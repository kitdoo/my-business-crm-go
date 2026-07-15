<script setup>
// Row click opens the ProductSKU view drawer (TD §12.4) rather than a
// movement-record view — same rationale as pages/inventory/index.vue.
const viewOverride = (item) => ({ entity: 'productSkus', id: item.skuId })
</script>

<template>
  <EntityListPage entity="inventoryMovements" :view-override="viewOverride" />
</template>
