<script setup>
// TD §12.1 tabbed detail page: General/Price/Movements. Inventory has no
// tab of its own — it can't be edited (stock only changes through
// InventoryMovementsService), so its "how much is where" summary lives at
// the top of the Movements tab instead (ProductStockSummary).
const route = useRoute()
const { t } = useI18n()

const tabItems = computed(() => [
  { label: t('entities.products.edit'), value: 'general' },
  { label: t('prices.tabLabel'), value: 'price' },
  { label: t('entities.inventoryMovements.label'), value: 'movements' },
])
const activeTab = ref('general')
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.products.edit') }}</h1>
    <UTabs :key="route.params.id" v-model="activeTab" :items="tabItems">
      <template #content="{ item }">
        <div class="pt-4">
          <ProductGeneralForm v-if="item.value === 'general'" :id="route.params.id" />
          <ProductPriceTab v-else-if="item.value === 'price'" :product-id="route.params.id" />
          <ProductMovementsTab v-else-if="item.value === 'movements'" :product-id="route.params.id" />
        </div>
      </template>
    </UTabs>
  </div>
</template>
