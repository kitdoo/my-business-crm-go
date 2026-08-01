<script setup>
// Plain quantity display for hundredths-of-a-unit fields (Inventory.quantity,
// InventoryMovement.quantity) — same basis-points convention as
// MoneyAmountLabel, via quantityAmount.js's toQuantity.
import { toQuantity } from '~/utils/quantityAmount.js'

const props = defineProps({
  value: { type: [Number, String], default: null },
})
const quantity = computed(() => (props.value == null ? '' : toQuantity(Number(props.value))))
</script>

<template>
  <span>{{ quantity }}</span>
</template>
