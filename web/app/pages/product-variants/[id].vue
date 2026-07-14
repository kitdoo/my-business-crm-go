<script setup>
// Tabbed detail page for one ProductVariant — same General/Price/
// Movements shape products/[id].vue used to have before sku/price/stock
// moved off Product onto ProductVariant.
const route = useRoute()
const { t } = useI18n()

const tabItems = computed(() => [
  { label: t('entities.productVariants.edit'), value: 'general' },
  { label: t('prices.tabLabel'), value: 'price' },
  { label: t('entities.inventoryMovements.label'), value: 'movements' },
])
const activeTab = ref('general')
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.productVariants.edit') }}</h1>
    <UTabs :key="route.params.id" v-model="activeTab" :items="tabItems">
      <template #content="{ item }">
        <div class="pt-4">
          <ProductVariantGeneralForm v-if="item.value === 'general'" :id="route.params.id" />
          <ProductPriceTab v-else-if="item.value === 'price'" :variant-id="route.params.id" />
          <ProductMovementsTab v-else-if="item.value === 'movements'" :variant-id="route.params.id" />
        </div>
      </template>
    </UTabs>
  </div>
</template>
