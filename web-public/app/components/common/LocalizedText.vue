<script setup>
// Read-only counterpart to LocalizedStringInput (TD §5.3). Never read
// `value.values[locale]` directly in a table/detail page — the fallback
// rule itself lives in utils/localizedText.js, shared with <RelationSelect>.
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  value: { type: Object, default: () => ({ values: {} }) },
})

const { locale } = useI18n()

const text = computed(() => localizedText(props.value, locale.value))
</script>

<template>
  <span>{{ text }}</span>
</template>
