<script setup>
// Product form followed by its variants, stacked on one page rather than
// split across tabs. Price/Movements moved off Product onto ProductVariant's
// own detail page (/product-variants/:id) — a product carries no sku/price/
// stock of its own, only variants do.
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
      <ProductVariantsTab :product-id="route.params.id" />
    </div>
  </div>
</template>
