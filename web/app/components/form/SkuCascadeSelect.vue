<script setup>
// Picks a ProductSKU id through Product -> Variant -> SKU (TD §9.3), the
// same chain the Sale line-item picker uses (pages/sales/new.vue) — a raw
// SKU code alone isn't recognizable, and ProductSKUsService.List has no
// free-text search (only exact skus/variantIds), so a flat searchable
// dropdown of every SKU isn't an option. Product is searchable (its
// backend Filter has searchQuery); Variant/SKU are scoped to their parent
// and stay small, so a plain list is fine there.
// Create-only: doesn't resolve an existing modelValue back into its
// product/variant, since every caller (currently just the InventoryMovement
// create form) only ever starts from a blank value.
const props = defineProps({
  modelValue: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const productId = ref(null)
const variantId = ref(null)

function onProductChange(id) {
  productId.value = id
  variantId.value = null
  emit('update:modelValue', null)
}
function onVariantChange(id) {
  variantId.value = id
  emit('update:modelValue', null)
}
function onSkuChange(id) {
  emit('update:modelValue', id)
}
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4 md:col-span-2">
    <RelationSelect
      :model-value="productId"
      relation="products"
      :label="t('fields.product')"
      required
      searchable
      @update:model-value="onProductChange"
    />
    <RelationSelect
      :model-value="variantId"
      relation="productVariants"
      :label="t('fields.variant')"
      :filter="{ productIds: [productId] }"
      :disabled="!productId"
      required
      @update:model-value="onVariantChange"
    />
    <RelationSelect
      :model-value="modelValue"
      relation="productSkus"
      :label="label || t('fields.sku')"
      :filter="{ variantIds: [variantId] }"
      :disabled="!variantId"
      :required="required"
      :error="error"
      @update:model-value="onSkuChange"
    />
  </div>
</template>
