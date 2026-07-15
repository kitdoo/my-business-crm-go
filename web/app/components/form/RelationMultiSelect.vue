<script setup>
// Select several related entities by id (TD §9.3, e.g. Product.categoryIds)
// — the multi-value sibling of <RelationSelect>. Loads the relation's full
// list once, same caveat about free-text search once lists grow large.
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  relation: { type: String, required: true }, // entity key, e.g. 'categories'
  modelValue: { type: Array, default: () => [] },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { locale } = useI18n()
const api = useEntityApi(props.relation)
const loading = ref(true)
const items = ref([])

function labelFor(item) {
  return typeof item.name === 'object' ? localizedText(item.name, locale.value) : item.name
}

async function load() {
  loading.value = true
  try {
    const res = await api.list({ pagination: { limit: 200 } })
    items.value = (res.items || []).map((item) => ({ label: labelFor(item), value: item.id }))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <UFormField :label="label" :error="error">
    <USelectMenu
      :model-value="modelValue"
      :items="items"
      :loading="loading"
      :disabled="disabled"
      multiple
      value-key="value"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
