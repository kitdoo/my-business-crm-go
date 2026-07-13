<script setup>
// Select a related entity by id (TD §9.3). Loads the relation's full list
// once (no free-text search yet — fine while lists stay small; add
// debounced search once an entity's list grows large enough to need it).
import { relationLabel } from '~/utils/relationLabel.js'

const props = defineProps({
  relation: { type: String, required: true }, // entity key, e.g. 'categories'
  modelValue: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { t, locale } = useI18n()
const api = useEntityApi(props.relation)
const loading = ref(true)
const optionItems = ref([])

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [...none, ...optionItems.value]
})

async function load() {
  loading.value = true
  try {
    const res = await api.list({ pagination: { limit: 200 } })
    optionItems.value = (res.items || []).map((item) => ({ label: relationLabel(item, locale.value), value: item.id }))
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <USelect
      :model-value="modelValue"
      :items="items"
      :loading="loading"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
