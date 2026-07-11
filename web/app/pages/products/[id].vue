<script setup>
// TD §12.1 describes a tabbed detail page (General/Price/Inventory/
// Movements) — the Inventory/Movements tabs land once those entities
// exist (tasks #10-#11); General + Price are wired up here.
const route = useRoute()
const { t } = useI18n()

const tabItems = computed(() => [
  { label: t('entities.products.edit'), value: 'general' },
  { label: t('prices.tabLabel'), value: 'price' },
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
        </div>
      </template>
    </UTabs>
  </div>
</template>
