<script setup>
// Select a related entity by id (TD §9.3). Loads the relation's full list
// once (no free-text search yet — fine while lists stay small; add
// debounced search once an entity's list grows large enough to need it).
// `tree: true` renders a self-referential list (e.g. Category.parentId)
// depth-indented in parent-before-children order instead of flat.
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  relation: { type: String, required: true }, // entity key, e.g. 'categories'
  modelValue: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
  excludeId: { type: String, default: null }, // hide this id from options (self-relation edit)
  tree: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { t, locale } = useI18n()
const api = useEntityApi(props.relation)
const loading = ref(true)
const optionItems = ref([])

function labelFor(item) {
  return typeof item.name === 'object' ? localizedText(item.name, locale.value) : item.name
}

// Depth-first, parent before children, roots first — a flat list a
// <USelect> can render with simple indentation (TD §12: "select с
// отступами по глубине").
function buildTreeOrder(all) {
  const byParent = new Map()
  for (const item of all) {
    const key = item.parentId || null
    if (!byParent.has(key)) byParent.set(key, [])
    byParent.get(key).push(item)
  }
  const ordered = []
  function walk(parentId, depth) {
    for (const item of byParent.get(parentId) || []) {
      ordered.push({ item, depth })
      walk(item.id, depth + 1)
    }
  }
  walk(null, 0)
  return ordered
}

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [...none, ...optionItems.value]
})

async function load() {
  loading.value = true
  try {
    const res = await api.list({ pagination: { limit: 200 } })
    const all = (res.items || []).filter((item) => item.id !== props.excludeId)
    const ordered = props.tree ? buildTreeOrder(all) : all.map((item) => ({ item, depth: 0 }))
    optionItems.value = ordered.map(({ item, depth }) => ({
      label: `${'— '.repeat(depth)}${labelFor(item)}`,
      value: item.id,
    }))
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
