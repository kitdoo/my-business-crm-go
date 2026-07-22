<script setup>
// SKUs section of a variant block (see ProductVariantsPanel.vue) — the
// bottom tier of the Product page's create/edit hierarchy. Each SKU is a
// collapsible block: collapsed shows sku code + total stock as a plain
// number (see fetchTotalStock — no per-warehouse table here, that level
// of detail lives on the dedicated /inventory-movements page), expanded
// shows its one compact General form (sku/price/attributes together,
// locked/edit pattern — see ProductSkuGeneralForm.vue). Saving a new SKU
// auto-expands it so the operator can see it settle in place.
import { fetchTotalStock, fetchTotalStockBySkuIds } from '~/utils/inventoryStock.js'

const props = defineProps({
  variantId: { type: String, required: true },
})

const { t } = useI18n()
const config = getEntityConfig('productSkus')
const api = useEntityApi('productSkus')
const inventoryApi = useEntityApi('inventory')
const { handle } = useApiErrorHandler()
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))

const items = ref([])
const stockBySkuId = ref({})
const loading = ref(true)
const creating = ref(false)
const expandedId = ref(null)

async function loadStock(skuItems) {
  stockBySkuId.value = await fetchTotalStockBySkuIds(
    inventoryApi,
    skuItems.map((item) => item.id),
  )
}

async function load() {
  loading.value = true
  try {
    const res = await api.list({
      filter: { variantIds: [props.variantId] },
      sort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_ASC' },
      pagination: { limit: 100 },
    })
    items.value = res.items || []
    await loadStock(items.value)
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

function toggle(item) {
  expandedId.value = expandedId.value === item.id ? null : item.id
}

async function onCreated(record) {
  creating.value = false
  items.value = [...items.value, record]
  expandedId.value = record.id
  stockBySkuId.value = { ...stockBySkuId.value, [record.id]: await fetchTotalStock(inventoryApi, record.id) }
}

function onDeleted(item) {
  items.value = items.value.filter((v) => v.id !== item.id)
  if (expandedId.value === item.id) expandedId.value = null
}

onMounted(load)
</script>

<template>
  <div class="space-y-3">
    <div v-if="loading" class="py-4 text-center text-sm text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <p v-if="items.length === 0 && !creating" class="text-sm text-neutral-500">{{ t('common.empty') }}</p>

      <div v-for="item in items" :key="item.id" class="rounded-lg border border-default">
        <button type="button" class="w-full flex items-center gap-3 p-3 text-left cursor-pointer" @click="toggle(item)">
          <span class="flex-1 text-sm font-mono">{{ item.sku }}</span>
          <span class="text-sm text-neutral-500">{{ t('entities.inventory.label') }}: {{ stockBySkuId[item.id] ?? 0 }}</span>
          <UIcon
            :name="expandedId === item.id ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
            class="size-4 text-neutral-400"
          />
        </button>

        <div v-if="expandedId === item.id" class="border-t border-default p-4">
          <ProductSkuGeneralForm :id="item.id" @deleted="onDeleted(item)" />
        </div>
      </div>
    </template>

    <div v-if="creating" class="rounded-lg border border-default p-4">
      <ProductSkuGeneralForm :variant-id="variantId" @created="onCreated" @cancel="creating = false" />
    </div>

    <UButton v-if="canCreate && !creating" icon="i-lucide-plus" color="neutral" variant="soft" @click="creating = true">
      {{ t('common.create') }}
    </UButton>
  </div>
</template>
