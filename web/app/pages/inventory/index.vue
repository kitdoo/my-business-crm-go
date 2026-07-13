<script setup>
// Row click opens the Product view drawer (TD §12.4) rather than an
// Inventory-record view — stock rows aren't independently interesting,
// the product they're for is.
const viewOverride = (item) => ({ entity: 'products', id: item.productId })
</script>

<template>
  <EntityListPage entity="inventory" :view-override="viewOverride" />
</template>
