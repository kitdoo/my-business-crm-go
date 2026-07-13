<script setup>
const props = defineProps({ period: { type: Object, required: true } })
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

// GetSalesByPartner returns every partner for the period with no
// limit/sort param — top 6 by amount, sorted client-side.
const topRows = computed(() => [...rows.value].sort((a, b) => Number(b.totalAmount) - Number(a.totalAmount)).slice(0, 6))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.salesByPartner(props.period)
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
  <DashboardWidget :title="t('dashboard.salesByPartner')" :loading="loading" :error="error" @retry="load">
    <div v-if="topRows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <table v-else class="w-full text-sm">
      <thead>
        <tr class="text-left text-neutral-500">
          <th class="py-1 pr-2 font-normal">{{ t('fields.partner') }}</th>
          <th class="py-1 pr-2 font-normal text-right">{{ t('dashboard.salesCount') }}</th>
          <th class="py-1 pr-2 font-normal text-right">{{ t('fields.totalAmount') }}</th>
          <th class="py-1 font-normal text-right">{{ t('dashboard.commission') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in topRows" :key="row.partnerId" class="border-b border-neutral-100 last:border-0">
          <td class="py-1 pr-2"><RelationLabel :value="row.partnerId" relation="partners" /></td>
          <td class="py-1 pr-2 text-right text-neutral-500">{{ row.salesCount }}</td>
          <td class="py-1 pr-2 text-right"><MoneyAmountLabel :value="row.totalAmount" /></td>
          <td class="py-1 text-right font-medium"><MoneyAmountLabel :value="row.commissionAmount" /></td>
        </tr>
      </tbody>
    </table>
  </DashboardWidget>
</template>
