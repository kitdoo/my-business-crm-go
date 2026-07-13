<script setup>
// Read-only latest-price line for the product view drawer — the editable
// counterpart lives in ProductPriceTab.vue (its own tab on the full detail
// page); this is deliberately just the top paragraph of that, no form, no
// history toggle.
const props = defineProps({
  productId: { type: String, required: true },
})

const { t } = useI18n()
const { formatMoney } = useFormatMoney()
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const loading = ref(true)
const price = ref(null)

async function load() {
  loading.value = true
  price.value = null
  try {
    price.value = await priceApi.get(props.productId)
  } catch (err) {
    if (err?.data?.error?.code !== 'NOT_FOUND') handle(err)
  } finally {
    loading.value = false
  }
}

watch(() => props.productId, load, { immediate: true })
</script>

<template>
  <div class="space-y-1">
    <h3 class="text-sm font-medium text-neutral-500">{{ t('prices.tabLabel') }}</h3>
    <div v-if="loading" class="py-2 text-sm text-neutral-500">{{ t('common.loading') }}</div>
    <p v-else-if="!price" class="text-sm text-neutral-500">{{ t('prices.noPriceYet') }}</p>
    <p v-else class="text-sm">
      {{ formatMoney(price.priceAmount, price.currency) }}
      <span v-if="price.discountAmount" class="text-neutral-500 ml-2">
        {{ t('prices.discount') }}: {{ formatMoney(price.discountAmount, price.currency) }}
      </span>
    </p>
  </div>
</template>
