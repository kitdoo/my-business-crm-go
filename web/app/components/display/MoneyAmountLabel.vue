<script setup>
// Plain amount display for basis-points fields that carry no currency
// code on their own type (Sale.totalAmount, SaleItem.priceAmount) —
// unlike ProductPrice, which has `currency` and uses useFormatMoney
// instead. TD §5.4's money formatting still applies (basis points / 100,
// never toFixed on a raw amount elsewhere), just without a currency
// symbol since the backend doesn't tell us which one applies here.
const props = defineProps({
  value: { type: [Number, String], default: null },
})
const amount = computed(() => (props.value == null ? '' : (Number(props.value) / 100).toFixed(2)))
</script>

<template>
  <span>{{ amount }}</span>
</template>
