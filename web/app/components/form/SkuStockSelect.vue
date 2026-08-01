<script setup>
// SKU picker for the Sale line-item form (pages/sales/new.vue), scoped to
// SKUs of the chosen variant that actually have stock somewhere (Inventory
// batch-looked-up via Filter.skuIds — see web/proto/inventory.proto) — a
// plain productSkus RelationSelect would offer SKUs with zero stock
// everywhere, which can never be sold. Mirrors WarehouseStockSelect's
// pattern (scope a picker by Inventory, surface the numbers to the
// caller) one level up the cascade.
// Emits the selected SKU's stock breakdown so the parent can cap the
// auto-mode quantity input and list per-warehouse availability for manual
// allocation, without a second inventory fetch.
import { relationLabel } from '~/utils/relationLabel.js'
import { toQuantity } from '~/utils/quantityAmount.js'

const props = defineProps({
  modelValue: { type: String, default: null },
  variantId: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue', 'update:stock'])

const { t, locale } = useI18n()
const productSkuApi = useEntityApi('productSkus')
const inventoryApi = useEntityApi('inventory')
const warehousesApi = useEntityApi('warehouses')

const loading = ref(false)
// One row per SKU of this variant that has stock in at least one
// warehouse: { value: skuId, label, totalStock, stock: [{warehouseId, quantity}] }.
const rows = ref([])

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [
    ...none,
    ...rows.value.map((r) => ({
      label: `${r.label} — ${t('entities.sales.inStock', { count: r.totalStock })}`,
      value: r.value,
    })),
  ]
})

async function load() {
  if (!props.variantId) {
    rows.value = []
    return
  }
  loading.value = true
  try {
    const skuRes = await productSkuApi.list({ filter: { variantIds: [props.variantId] }, pagination: { limit: 200 } })
    const skus = skuRes.items || []
    if (!skus.length) {
      rows.value = []
      return
    }
    const [invRes, warehousesRes] = await Promise.all([
      inventoryApi.list({ filter: { skuIds: skus.map((s) => s.id) }, pagination: { limit: 500 } }),
      warehousesApi.list({ pagination: { limit: 200 } }),
    ])
    const warehouseById = new Map((warehousesRes.items || []).map((w) => [w.id, w]))
    const stockBySku = new Map()
    for (const inv of invRes.items || []) {
      if (!(Number(inv.quantity) > 0)) continue
      const warehouse = warehouseById.get(inv.warehouseId)
      const list = stockBySku.get(inv.skuId) || []
      list.push({
        warehouseId: inv.warehouseId,
        warehouseLabel: warehouse ? relationLabel(warehouse, locale.value) : inv.warehouseId,
        quantity: toQuantity(Number(inv.quantity)),
      })
      stockBySku.set(inv.skuId, list)
    }
    rows.value = skus
      .map((sku) => {
        const stock = stockBySku.get(sku.id) || []
        return {
          value: sku.id,
          label: relationLabel(sku, locale.value),
          stock,
          totalStock: stock.reduce((sum, s) => sum + s.quantity, 0),
        }
      })
      .filter((r) => r.totalStock > 0)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.variantId, () => {
  emit('update:modelValue', null)
  load()
})

watch(
  () => [props.modelValue, rows.value],
  () => {
    const row = rows.value.find((r) => r.value === props.modelValue)
    emit('update:stock', row ? { stock: row.stock, totalStock: row.totalStock } : { stock: [], totalStock: null })
  },
  { deep: true },
)
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <USelectMenu
      :model-value="modelValue"
      :items="items"
      :loading="loading"
      :disabled="!variantId"
      value-key="value"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
