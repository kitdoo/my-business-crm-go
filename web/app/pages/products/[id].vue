<script setup>
// The entire product creation/editing workflow lives on this one page:
// Product -> Variants -> SKUs -> Price/Stock, all stacked inline, no tabs
// and no navigating away to a separate page for any of it (see
// ProductVariantsPanel.vue / ProductSkusPanel.vue). A product carries no
// sku/price/stock of its own, only variants (and their SKUs) do.
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <div class="space-y-8">
    <div class="space-y-4">
      <h1 class="text-xl font-semibold">{{ t('entities.products.edit') }}</h1>
      <ProductGeneralForm :id="route.params.id" />
    </div>

    <div class="space-y-4">
      <h2 class="text-lg font-semibold">{{ t('entities.productVariants.label') }}</h2>
      <ProductVariantsPanel :product-id="route.params.id" />
    </div>
  </div>
</template>
