<script setup>
// Read-only "id -> display name" for a related entity (TD §9.3) — resolves
// through useReferenceCacheStore so repeated ids across a table batch into
// one request instead of one Get per row (TD §9.5).
import { localizedText } from '~/utils/localizedText.js'
import { useReferenceCacheStore } from '~/stores/referenceCache'

const props = defineProps({
  value: { type: String, default: null }, // related entity id
  relation: { type: String, required: true }, // entity key, e.g. 'categories'
})

const { locale } = useI18n()
const store = useReferenceCacheStore()
const item = ref(props.value ? store.getCached(props.relation, props.value) : null)

watchEffect(async () => {
  item.value = props.value ? await store.resolve(props.relation, props.value) : null
})

const text = computed(() => {
  if (!item.value) return ''
  const name = item.value.name
  return typeof name === 'object' ? localizedText(name, locale.value) : name
})
</script>

<template>
  <span>{{ text }}</span>
</template>
