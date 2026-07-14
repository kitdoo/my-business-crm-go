<script setup>
// Row click opens the ProductVariant view drawer (TD §12.4) rather than
// an Inventory-record view — stock rows aren't independently interesting,
// the variant they're for is.
const viewOverride = (item) => ({ entity: 'productVariants', id: item.variantId })
</script>

<template>
  <EntityListPage entity="inventory" :view-override="viewOverride" />
</template>
