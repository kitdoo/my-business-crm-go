<script setup>
// The dashboard's one headline number: total sales amount for the period
// (replaces the former separate Turnover + sales-count/total tiles —
// TD §8.4 update, one focused figure instead of three).
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const totalAmount = ref(0)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.salesReport(props.period)
    totalAmount.value = res.rows.reduce((sum, row) => sum + Number(row.totalAmount || 0), 0)
  } catch (err) {
    error.value = err?.data?.error?.message || t('errors.generic')
  } finally {
    loading.value = false
  }
}

watch(() => props.period, load, { immediate: true })
</script>

<template>
  <StatTile :label="t('dashboard.salesAmount')" :loading="loading" :error="error">
    <MoneyAmountLabel :value="totalAmount" />
  </StatTile>
</template>
