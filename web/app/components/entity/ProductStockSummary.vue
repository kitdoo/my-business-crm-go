<script setup>
// Read-only "how much is where" summary for a product variant (TD
// §12.1/§12.4) — Inventory has no Create/Update/Delete permissions at all
// (stock only changes through InventoryMovementsService), so there is
// nothing to edit here. Shown in the variant view drawer
// (EntityViewDrawer.vue) via config.view.priceStockSummary.
const props = defineProps({
  variantId: { type: String, required: true },
})

const { t } = useI18n()
const api = useEntityApi('inventory')
const loading = ref(true)
const items = ref([])

async function load() {
  loading.value = true
  try {
    const res = await api.list({ filter: { variantId: props.variantId }, pagination: { limit: 200 } })
    items.value = res.items || []
  } finally {
    loading.value = false
  }
}

watch(() => props.variantId, load, { immediate: true })
</script>

<template>
  <div class="space-y-2">
    <h3 class="text-sm font-medium text-neutral-500">{{ t('entities.inventory.label') }}</h3>
    <div v-if="loading" class="py-2 text-sm text-neutral-500">{{ t('common.loading') }}</div>
    <p v-else-if="!items.length" class="text-sm text-neutral-500">{{ t('common.empty') }}</p>
    <ul v-else class="divide-y divide-neutral-100 rounded-md border border-neutral-100">
      <li v-for="item in items" :key="item.id" class="flex items-center justify-between px-3 py-2 text-sm">
        <RelationLabel :value="item.warehouseId" relation="warehouses" />
        <span class="font-medium">{{ item.quantity }}</span>
      </li>
    </ul>
  </div>
</template>
