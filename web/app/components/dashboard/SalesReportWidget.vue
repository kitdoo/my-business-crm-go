<script setup>
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

const totals = computed(() => ({
  salesCount: rows.value.reduce((sum, row) => sum + Number(row.salesCount || 0), 0),
  totalAmount: rows.value.reduce((sum, row) => sum + Number(row.totalAmount || 0), 0),
}))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.salesReport(props.period)
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
  <DashboardWidget :title="t('dashboard.salesReport')" :loading="loading" :error="error" @retry="load">
    <div class="flex justify-between text-sm">
      <span class="text-neutral-500">{{ t('dashboard.salesCount') }}</span>
      <span class="font-medium">{{ totals.salesCount }}</span>
    </div>
    <div class="flex justify-between text-sm">
      <span class="text-neutral-500">{{ t('fields.totalAmount') }}</span>
      <span class="font-medium"><MoneyAmountLabel :value="totals.totalAmount" /></span>
    </div>
  </DashboardWidget>
</template>
