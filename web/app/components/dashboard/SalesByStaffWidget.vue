<script setup>
// userId isn't resolved via <RelationLabel relation="users"> here because
// Users isn't necessarily readable by every role that can see reports
// (admin-only entity, TD §10); showing raw ids is an accepted limitation
// per TD §8.4 rather than gating the whole widget on users:read too.
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
    <div v-if="rows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <table v-else class="w-full text-sm">
      <tbody>
        <tr v-for="row in rows" :key="row.userId" class="border-b border-neutral-100 last:border-0">
          <td class="py-1 pr-2 font-mono text-xs text-neutral-500">{{ row.userId }}</td>
          <td class="py-1 pr-2 text-right text-neutral-500">{{ row.salesCount }}</td>
          <td class="py-1 text-right font-medium"><MoneyAmountLabel :value="row.totalAmount" /></td>
        </tr>
      </tbody>
    </table>
  </DashboardWidget>
</template>
