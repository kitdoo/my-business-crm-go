<script setup>
// Dashboard (TD §8.4): all 6 report widgets, each loading independently.
// Hidden entirely (not per-widget) for a role without reports:read —
// all 6 ReportsService methods map to the same permission on the
// backend, so one check covers the whole page.
const { t } = useI18n()
const { can } = usePermission()

const period = ref({})
const canViewReports = computed(() => can('reports:read'))
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ t('nav.dashboard') }}</h1>
      <PeriodFilter v-if="canViewReports" v-model="period" />
    </div>

    <div v-if="canViewReports && period.from" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <TurnoverWidget :period="period" />
      <SalesReportWidget :period="period" />
      <PopularProductsWidget :period="period" />
      <SalesByStaffWidget :period="period" />
      <SalesByPartnerWidget :period="period" />
      <StockLevelsWidget />
    </div>

    <p v-else-if="!canViewReports" class="text-sm text-neutral-500">{{ t('dashboard.noAccess') }}</p>
  </div>
</template>
