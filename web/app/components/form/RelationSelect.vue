<script setup>
// Select a related entity by id (TD §9.3). By default loads the
// relation's full list once (fine while lists stay small — e.g. lines
// already scoped to a parent, like SKUs of one Variant). Set `searchable`
// for relations whose backend Filter has a `searchQuery` field (Brand,
// Category, Client, Partner, Product, User, Warehouse per
// PROTO_DEVELOPMENT_STANDARD.md's searchQuery convention) — the dropdown
// then debounces keystrokes into server-side searches instead of relying
// on the first 200 rows, which stops finding anything once the relation's
// table outgrows one page.
// `filter` scopes the loaded list to a dependent parent (e.g. SKUs of one
// Variant, in the Sale line-item picker) — reloads whenever it changes;
// pass `disabled` alongside it while the parent hasn't been picked yet, so
// the field doesn't briefly show every row of an unscoped list. `disabled`
// is also reused to lock an *already-picked* value read-only (product/
// variant/sku "locked" forms) — there the list still needs to load so the
// current id resolves to a label instead of showing raw, so the
// list-load only skips when both disabled AND nothing is picked yet.
import { relationLabel } from '~/utils/relationLabel.js'

const props = defineProps({
  relation: { type: String, required: true }, // entity key, e.g. 'categories'
  modelValue: { type: String, default: null },
  label: { type: String, default: '' },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
  filter: { type: Object, default: null },
  disabled: { type: Boolean, default: false },
  searchable: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

const { t, locale } = useI18n()
const api = useEntityApi(props.relation)
const loading = ref(true)
const optionItems = ref([])
const searchTerm = ref('')

const items = computed(() => {
  const none = props.required ? [] : [{ label: t('common.none'), value: null }]
  return [...none, ...optionItems.value]
})

async function load(query = '') {
  if (props.disabled && !props.modelValue) {
    optionItems.value = []
    loading.value = false
    return
  }
  loading.value = true
  try {
    const filter = { ...(props.filter || {}) }
    if (props.searchable && query) filter.searchQuery = query
    const res = await api.list({
      filter: Object.keys(filter).length ? filter : undefined,
      pagination: { limit: 200 },
    })
    optionItems.value = (res.items || []).map((item) => ({ label: relationLabel(item, locale.value), value: item.id }))
  } finally {
    loading.value = false
  }
}

onMounted(() => load())
watch(() => [props.filter, props.disabled], () => load(searchTerm.value), { deep: true })
// modelValue is watched separately, and only reloads while disabled: on a
// locked/edit form this component mounts before useEntityForm's async
// load() has populated the parent's form, so modelValue arrives a tick
// later — exactly when a disabled field needs to (re)fetch to resolve its
// label (see the guard above). Folding modelValue into the watch above
// instead would reload the whole list on every ordinary selection too
// (props.modelValue changing right back down after the user's own
// @update:model-value), which disabled-gating here avoids.
watch(
  () => props.modelValue,
  (value) => {
    if (props.disabled && value) load(searchTerm.value)
  },
)

// Debounced server-side search — a no-op watcher when !searchable, so
// non-searchable relations behave exactly as before (USelectMenu's own
// client-side filter just runs over the loaded page, same as USelect did).
let searchDebounce = null
watch(searchTerm, (term) => {
  if (!props.searchable) return
  clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => load(term), 300)
})
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <USelectMenu
      :model-value="modelValue"
      v-model:search-term="searchTerm"
      :items="items"
      :loading="loading"
      :disabled="disabled"
      :ignore-filter="searchable"
      value-key="value"
      class="w-full"
      @update:model-value="(v) => emit('update:modelValue', v)"
    />
  </UFormField>
</template>
