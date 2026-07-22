<script setup>
// Warehouse picker for the InventoryMovement create form, scoped to
// warehouses that actually carry the selected SKU (Inventory.List filtered
// by skuId) — a plain `warehouses` RelationSelect would offer every
// warehouse, most of which never stocked this SKU at all. Disabled until a
// SKU is chosen, same lockstep pattern as SkuCascadeSelect's own
// Product -> Variant -> SKU chain.
// Also surfaces the chosen warehouse's current on-hand quantity via
// `update:available`, so the sibling quantity field can cap how much can
// be taken out.
import { relationLabel } from '~/utils/relationLabel.js'

const props = defineProps({
  modelValue: { type: String, default: null },
  skuId: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
  // Hide warehouses whose current on-hand quantity for this SKU is zero —
  // InventoryMovement (adjustments/write-offs) leaves this off since a
  // warehouse can legitimately need a movement at zero stock; Sale (can
  // only ever draw from stock that exists) turns it on.
  onlyInStock: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue', 'update:available'])

const { t, locale } = useI18n()
const inventoryApi = useEntityApi('inventory')
const warehousesApi = useEntityApi('warehouses')

const loading = ref(false)
// One row per warehouse that has an Inventory record for this SKU:
// { value: warehouseId, label, quantity }.
const rows = ref([])

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [...none, ...rows.value.map((r) => ({ label: r.label, value: r.value }))]
})

async function load() {
  if (!props.skuId) {
    rows.value = []
    return
  }
  loading.value = true
  try {
    const [inventoryRes, warehousesRes] = await Promise.all([
      inventoryApi.list({ filter: { skuId: props.skuId }, pagination: { limit: 200 } }),
      warehousesApi.list({ pagination: { limit: 200 } }),
    ])
    const warehouseById = new Map((warehousesRes.items || []).map((w) => [w.id, w]))
    rows.value = (inventoryRes.items || [])
      .filter((i) => warehouseById.has(i.warehouseId))
      .map((i) => ({
        value: i.warehouseId,
        label: relationLabel(warehouseById.get(i.warehouseId), locale.value),
        quantity: Number(i.quantity),
      }))
      .filter((r) => !props.onlyInStock || r.quantity > 0)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.skuId, () => {
  emit('update:modelValue', null)
  load()
})

watch(
  () => [props.modelValue, rows.value],
  () => {
    const row = rows.value.find((r) => r.value === props.modelValue)
    emit('update:available', row ? row.quantity : null)
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
      :disabled="!skuId"
      value-key="value"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
