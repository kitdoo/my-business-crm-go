<script setup>
// One field of EntityForm.vue's config-driven form (TD §9.1/§9.3), factored
// out so the same per-type rendering can be shared between the main field
// grid and the "more information" disclosure (config.form.fields entries
// with `advanced: true`) without duplicating this v-else-if chain twice.
import { getEnumOptions } from '~/config/enums.js'

const props = defineProps({
  field: { type: Object, required: true },
  form: { type: Object, required: true }, // mutated directly by key, not replaced
  error: { type: String, default: undefined },
  isCreate: { type: Boolean, required: true },
  locales: { type: Array, required: true },
  // On-hand quantity for the currently picked skuId/warehouseId pair
  // (only meaningful for InventoryMovement.quantity, which has
  // capByStock) — passed down so this field's UInputNumber can be capped
  // by a sibling WarehouseStockSelect's reported availability.
  stockAvailable: { type: Number, default: null },
})
const emit = defineEmits(['update:available'])

const { t } = useI18n()

function quantityMin() {
  if (!props.field.capByStock || props.stockAvailable == null) return props.field.min
  return -props.stockAvailable
}

// field.scale (e.g. 100 for InventoryMovement.quantity, hundredths-of-a-
// unit on the wire — see quantityAmount.js) lets a 'number' field store its
// wire value in `form[field.key]` while the input itself shows/accepts the
// plain scaled-down number the operator actually types. Fields without a
// scale behave exactly as before (identity divide/multiply by 1).
const displayValue = computed(() => {
  const raw = props.form[props.field.key]
  if (!props.field.scale || raw == null) return raw
  return raw / props.field.scale
})
function setDisplayValue(value) {
  props.form[props.field.key] = !props.field.scale || value == null ? value : Math.round(value * props.field.scale)
}
</script>

<template>
  <LocalizedStringInput
    v-if="field.type === 'localizedString'"
    v-model="form[field.key]"
    :locales="locales"
    :required-locale="'sr'"
    :label="t(field.label)"
    :error="error"
    :required="field.required"
  />
  <UFormField v-else-if="field.type === 'enum'" :label="t(field.label)" :error="error">
    <USelect
      v-model="form[field.key]"
      :items="
        getEnumOptions(field.enum)
          .filter((v) => !field.excludeOptions?.includes(v))
          .map((v) => ({ label: t(`enums.status.${v}`), value: v }))
      "
      class="w-full"
    />
  </UFormField>
  <UFormField
    v-else-if="field.type === 'text'"
    :label="t(field.label)"
    :required="field.required"
    :error="error"
    :class="field.multiline ? 'md:col-span-2' : undefined"
  >
    <UTextarea
      v-if="field.multiline"
      v-model="form[field.key]"
      class="w-full"
      :required="field.required"
      :maxlength="field.maxLength"
    />
    <UInput
      v-else
      v-model="form[field.key]"
      class="w-full"
      :type="field.inputType || 'text'"
      :required="field.required"
      :maxlength="field.maxLength"
      :readonly="field.immutableOnEdit && !isCreate"
    />
  </UFormField>
  <UFormField
    v-else-if="field.type === 'number'"
    :label="t(field.label)"
    :required="field.required"
    :error="error"
    :hint="field.hint ? t(field.hint) : undefined"
    :class="field.fullWidth ? 'md:col-span-2' : undefined"
  >
    <UInputNumber
      :model-value="displayValue"
      class="w-full"
      :min="quantityMin()"
      :max="field.max"
      :step="field.scale ? 1 / field.scale : undefined"
      :disabled="field.requiresFields?.some((k) => !form[k])"
      @update:model-value="setDisplayValue"
    />
  </UFormField>
  <SkuCascadeSelect
    v-else-if="field.type === 'skuCascade'"
    v-model="form[field.key]"
    :label="t(field.label)"
    :error="error"
    :required="field.required"
  />
  <WarehouseStockSelect
    v-else-if="field.type === 'warehouseStock'"
    v-model="form[field.key]"
    :sku-id="form.skuId"
    :label="t(field.label)"
    :error="error"
    :required="field.required"
    @update:available="emit('update:available', $event)"
  />
  <RelationSelect
    v-else-if="field.type === 'relation'"
    v-model="form[field.key]"
    :relation="field.relation"
    :label="t(field.label)"
    :error="error"
    :required="field.required"
    :searchable="field.searchable"
  />
  <RelationMultiSelect
    v-else-if="field.type === 'relationMulti'"
    v-model="form[field.key]"
    :relation="field.relation"
    :label="t(field.label)"
    :error="error"
  />
  <UFormField v-else-if="field.type === 'images'" :label="t(field.label)" class="md:col-span-2">
    <ProductImageUploader v-model="form[field.key]" />
  </UFormField>
  <UFormField v-else-if="field.type === 'boolean'" :error="error">
    <UCheckbox v-model="form[field.key]" :label="t(field.label)" />
  </UFormField>
</template>
