<script setup>
// Clickable "#<number>" link to the sale that caused this row (e.g. an
// InventoryMovement's saleId) — renders nothing when unset, matching
// RelationLabel's resolution (reference-cache, TD §9.5) but linking to
// the sale's full page instead of just showing text.
import { relationLabel } from '~/utils/relationLabel.js'
import { useReferenceCacheStore } from '~/stores/referenceCache'

const props = defineProps({
  value: { type: String, default: null }, // saleId
})

const store = useReferenceCacheStore()
const item = ref(props.value ? store.getCached('sales', props.value) : null)

watchEffect(async () => {
  item.value = props.value ? await store.resolve('sales', props.value) : null
})

const text = computed(() => (item.value ? relationLabel(item.value) : ''))
</script>

<template>
  <NuxtLink v-if="value && item" :to="`/sales/${value}`" class="text-brand-700 hover:underline" @click.stop>
    {{ text }}
  </NuxtLink>
</template>
