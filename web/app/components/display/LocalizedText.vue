<script setup>
// Read-only counterpart to LocalizedStringInput (TD §5.3) — the one place
// the current-locale -> sr fallback rule lives. Never read
// `value.values[locale]` directly in a table/detail page.
const props = defineProps({
  value: { type: Object, default: () => ({ values: {} }) },
})

const { locale } = useI18n()

const text = computed(() => {
  const values = props.value?.values || {}
  return values[locale.value] ?? values.sr ?? ''
})
</script>

<template>
  <span>{{ text }}</span>
</template>
