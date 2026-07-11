<script setup>
const props = defineProps({
  amount: { type: [Number, String], default: null },
  currency: { type: String, default: '' },
})
const { formatMoney } = useFormatMoney()
const text = computed(() => formatMoney(Number(props.amount), props.currency))
</script>

<template>
  <span>{{ text }}</span>
</template>
