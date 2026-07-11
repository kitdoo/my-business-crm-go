<script setup>
// No period filter — stock levels are a current snapshot, not a range.
// Sorted client-side ascending (lowest stock first — "requires
// attention", TD §8.4) and capped to the top 10; the RPC itself has no
// limit param.
const { t } = useI18n()
const reportApi = useReportApi()

const loading = ref(true)
const error = ref('')
const rows = ref([])

const topRows = computed(() => [...rows.value].sort((a, b) => Number(a.quantity) - Number(b.quantity)).slice(0, 10))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await reportApi.stockLevels()
    rows.value = res.rows
  } catch (err) {
    error.value = err?.data?.error?.message || t('errors.generic')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <DashboardWidget :title="t('dashboard.stockLevels')" :loading="loading" :error="error" @retry="load">
    <div v-if="topRows.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</div>
    <template v-else>
      <table class="w-full text-sm">
        <tbody>
          <tr v-for="(row, index) in topRows" :key="index" class="border-b border-neutral-100 last:border-0">
            <td class="py-1 pr-2"><RelationLabel :value="row.productId" relation="products" /></td>
            <td class="py-1 pr-2 text-neutral-500"><RelationLabel :value="row.warehouseId" relation="warehouses" /></td>
            <td class="py-1 text-right font-medium">{{ row.quantity }}</td>
          </tr>
        </tbody>
      </table>
      <NuxtLink to="/inventory" class="text-sm text-brand-700 hover:underline">{{ t('dashboard.viewAllInventory') }}</NuxtLink>
    </template>
  </DashboardWidget>
</template>
