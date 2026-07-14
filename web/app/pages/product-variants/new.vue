<script setup>
// productId arrives as a query param (see ProductVariantsTab.vue /
// EntityListPage's createDefaults) — there's no other way to hand a
// value to a full create page.
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.productVariants.create') }}</h1>
    <ProductVariantGeneralForm :id="null" :product-id="route.query.productId?.toString()" />
  </div>
</template>
