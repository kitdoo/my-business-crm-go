<script setup>
const props = defineProps({
  amount: { type: [Number, String], default: null },
  currency: { type: String, default: '' },
  // Product.priceUnit, e.g. 'PRICE_UNIT_SQUARE_METER' — rendered as a
  // "/ unit" suffix so a per-m² price is never mistaken for a per-piece
  // one (see tmp/CENA PHOMI MATERIJALA.pdf mixup this was added to fix).
  unit: { type: String, default: '' },
})
const { t } = useI18n()
const { formatMoney } = useFormatMoney()
const text = computed(() => formatMoney(Number(props.amount), props.currency))
const unitLabel = computed(() => {
  if (props.unit === 'PRICE_UNIT_SQUARE_METER') return t('catalog.priceUnit.squareMeter')
  if (props.unit === 'PRICE_UNIT_PIECE') return t('catalog.priceUnit.piece')
  return ''
})
</script>

<template>
  <span>{{ text }}<span v-if="unitLabel" class="text-black/50 font-normal"> / {{ unitLabel }}</span></span>
</template>
