<script setup>
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.popularProducts(props.period, 10)
    rows.value = res.rows
  } catch (err) {
    error.value = err?.data?.error?.message || t('errors.generic')
  } finally {
    loading.value = false
  }
}

watch(() => props.period, load, { immediate: true })
</script>

<template>
  <DashboardWidget :title="t('dashboard.popularProducts')" :loading="loading" :error="error" @retry="load">
    <div v-if="rows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <table v-else class="w-full text-sm">
      <tbody>
        <tr v-for="row in rows" :key="row.variantId" class="border-b border-neutral-100 last:border-0">
          <td class="py-1 pr-2"><RelationLabel :value="row.variantId" relation="productVariants" /></td>
          <td class="py-1 pr-2 text-right text-neutral-500">{{ row.quantitySold }}</td>
          <td class="py-1 text-right font-medium"><MoneyAmountLabel :value="row.totalAmount" /></td>
        </tr>
      </tbody>
    </table>
  </DashboardWidget>
</template>
