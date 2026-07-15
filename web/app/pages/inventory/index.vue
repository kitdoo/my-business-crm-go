<script setup>
// Row click opens the ProductSKU view drawer (TD §12.4) rather than
// an Inventory-record view — stock rows aren't independently interesting,
// the SKU they're for is.
const viewOverride = (item) => ({ entity: 'productSkus', id: item.skuId })
</script>

<template>
  <EntityListPage entity="inventory" :view-override="viewOverride" />
</template>
