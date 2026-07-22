<script setup>
// Variant picker for the Sale line-item form (pages/sales/new.vue), scoped
// to variants of the chosen product that have at least one SKU with stock
// somewhere. A plain productVariants RelationSelect lists every variant
// regardless of stock — picking one with none was a dead end: SkuStockSelect
// one level below would come back empty and there was no way to continue
// that line item. Mirrors SkuStockSelect's own pattern (scope a picker by
// Inventory) one level up the cascade.
import { relationLabel } from '~/utils/relationLabel.js'

const props = defineProps({
  modelValue: { type: String, default: null },
  productId: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { t, locale } = useI18n()
const variantApi = useEntityApi('productVariants')
const productSkuApi = useEntityApi('productSkus')
const inventoryApi = useEntityApi('inventory')

const loading = ref(false)
// One row per variant of this product with at least one in-stock SKU.
const rows = ref([])

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [...none, ...rows.value]
})

async function load() {
  if (!props.productId) {
    rows.value = []
    return
  }
  loading.value = true
  try {
    const variantRes = await variantApi.list({ filter: { productIds: [props.productId] }, pagination: { limit: 200 } })
    const variants = variantRes.items || []
    if (!variants.length) {
      rows.value = []
      return
    }
    const skuRes = await productSkuApi.list({
      filter: { variantIds: variants.map((v) => v.id) },
      pagination: { limit: 500 },
    })
    const skus = skuRes.items || []
    if (!skus.length) {
      rows.value = []
      return
    }
    const invRes = await inventoryApi.list({ filter: { skuIds: skus.map((s) => s.id) }, pagination: { limit: 1000 } })
    const stockedSkuIds = new Set(
      (invRes.items || []).filter((inv) => Number(inv.quantity) > 0).map((inv) => inv.skuId),
    )
    const stockedVariantIds = new Set(skus.filter((s) => stockedSkuIds.has(s.id)).map((s) => s.variantId))
    rows.value = variants
      .filter((v) => stockedVariantIds.has(v.id))
      .map((v) => ({ value: v.id, label: relationLabel(v, locale.value) }))
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.productId, () => {
  emit('update:modelValue', null)
  load()
})
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <USelectMenu
      :model-value="modelValue"
      :items="items"
      :loading="loading"
      :disabled="!productId"
      value-key="value"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
