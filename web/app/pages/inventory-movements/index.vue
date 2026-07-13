<script setup>
// Row click opens the Product view drawer (TD §12.4) rather than a
// movement-record view — same rationale as pages/inventory/index.vue.
const viewOverride = (item) => ({ entity: 'products', id: item.productId })
</script>

<template>
  <EntityListPage entity="inventoryMovements" :view-override="viewOverride" />
</template>
