<script setup>
const props = defineProps({
  value: { type: [Number, String], default: null },
  style: { type: String, default: 'short' },
})
const { formatDate } = useFormatDate()
const seconds = computed(() => (props.value != null ? Number(props.value) : null))
</script>

<template>
  <span>{{ formatDate(seconds, style) }}</span>
</template>
