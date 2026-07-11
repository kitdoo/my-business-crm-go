<script setup>
// Plain-text i18n label for an enum value that isn't a status (no color
// badge needed — e.g. User.role). Shares the same flat `enums.status.*`
// i18n namespace as <StatusBadge> (TD §7): every enum value string is
// globally unique across the project, so one namespace is enough.
const props = defineProps({
  value: { type: String, default: '' },
})
const { t } = useI18n()
const text = computed(() => t(`enums.status.${props.value}`, props.value))
</script>

<template>
  <span>{{ text }}</span>
</template>
