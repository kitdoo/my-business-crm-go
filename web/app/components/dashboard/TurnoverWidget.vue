<script setup>
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const { formatDate } = useFormatDate()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.turnover(props.period)
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
  <DashboardWidget :title="t('dashboard.turnover')" :loading="loading" :error="error" @retry="load">
    <div v-if="rows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <table v-else class="w-full text-sm">
      <tbody>
        <tr v-for="(row, index) in rows" :key="index" class="border-b border-neutral-100 last:border-0">
          <td class="py-1 pr-2 text-neutral-500">{{ formatDate(row.periodStart) }} — {{ formatDate(row.periodEnd) }}</td>
          <td class="py-1 text-right font-medium"><MoneyAmountLabel :value="row.totalAmount" /></td>
        </tr>
      </tbody>
    </table>
  </DashboardWidget>
</template>
