<script setup>
// userId resolves via <RelationLabel relation="users">; a viewer without
// users:read still degrades gracefully to a blank label (existing
// RelationLabel/reference-cache behavior) rather than blocking the whole
// widget — same accepted limitation as before, just a real name for
// whoever can see one (mainly admins, who are the primary dashboard
// audience).
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

// GetSalesByStaff returns every user for the period with no limit/sort
// param — top 6 by amount, sorted client-side (same pattern the removed
// StockLevelsWidget used for its top-10 list).
const topRows = computed(() => [...rows.value].sort((a, b) => Number(b.totalAmount) - Number(a.totalAmount)).slice(0, 6))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.salesByStaff(props.period)
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
  <DashboardWidget :title="t('dashboard.salesByStaff')" :loading="loading" :error="error" @retry="load">
    <div v-if="topRows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <table v-else class="w-full text-sm">
      <tbody>
        <tr v-for="row in topRows" :key="row.userId" class="border-b border-neutral-100 last:border-0">
          <td class="py-1 pr-2"><RelationLabel :value="row.userId" relation="users" /></td>
          <td class="py-1 pr-2 text-right text-neutral-500">{{ row.salesCount }}</td>
          <td class="py-1 text-right font-medium"><MoneyAmountLabel :value="row.totalAmount" /></td>
        </tr>
      </tbody>
    </table>
  </DashboardWidget>
</template>
