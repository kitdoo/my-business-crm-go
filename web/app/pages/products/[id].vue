<script setup>
// TD §12.1 describes a tabbed detail page (General/Price/Inventory/
// Movements) — the Movements tab lands once InventoryMovements exists
// (task #11).
const route = useRoute()
const { t } = useI18n()

const tabItems = computed(() => [
  { label: t('entities.products.edit'), value: 'general' },
  { label: t('prices.tabLabel'), value: 'price' },
  { label: t('entities.inventory.label'), value: 'inventory' },
])
const activeTab = ref('general')
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.products.edit') }}</h1>
    <UTabs v-model="activeTab" :items="tabItems">
      <template #content="{ item }">
        <div class="pt-4">
          <EntityForm v-if="item.value === 'general'" entity="products" :id="route.params.id" />
          <ProductPriceTab v-else-if="item.value === 'price'" :product-id="route.params.id" />
          <EntityDataTable v-else-if="item.value === 'inventory'" entity="inventory" :fixed-filter="{ productId: route.params.id }" />
        </div>
      </template>
    </UTabs>
  </div>
</template>
