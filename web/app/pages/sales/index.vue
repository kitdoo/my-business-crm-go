<script setup>
// Not <EntityListPage> — Sale has no generic form/drawer (TD §12.3), so
// "Create" is a link to the wizard page and rows navigate to the detail
// page instead of opening a drawer. The row-actions icons below open
// SaleInvoiceModal/SaleCancelModal in place (see those components) rather
// than navigating to the detail page first — a click used to both fire
// the action *and* leave the list, which is the opposite of what an
// inline row action should do.
import { isSaleTerminal } from '~/utils/saleStatus.js'

const config = getEntityConfig('sales')
const { t } = useI18n()
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))
const canUpdate = computed(() => can('sales:update'))
const canGenerateInvoice = computed(() => can('invoices:generate'))
const canGenerateSalesReport = computed(() => can('salesreport:generate'))

const invoiceModal = ref(null)
const cancelModal = ref(null)
const table = ref(null)
const generatingReport = ref(false)
const reportError = ref('')

const reportPeriods = [
  { months: 1, label: 'entities.sales.reportPeriod1m' },
  { months: 3, label: 'entities.sales.reportPeriod3m' },
  { months: 6, label: 'entities.sales.reportPeriod6m' },
  { months: 12, label: 'entities.sales.reportPeriod1y' },
]
const reportItems = computed(() => [
  reportPeriods.map((p) => ({ label: t(p.label), onSelect: () => generateReport(p.months) })),
])

// Backend takes an inclusive Unix-second [from, to] range (see
// entities.SalesReportGenerate) — "N months" is computed client-side off
// the current moment rather than a fixed calendar window, so "last 1
// month" always means the 30-ish days up to now.
async function generateReport(months) {
  generatingReport.value = true
  reportError.value = ''
  try {
    const to = new Date()
    const from = new Date(to)
    from.setMonth(from.getMonth() - months)

    const blob = await $fetch('/api/sales-report', {
      method: 'POST',
      body: { from: Math.floor(from.getTime() / 1000), to: Math.floor(to.getTime() / 1000) },
      responseType: 'blob',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `sales-report-${to.toISOString().slice(0, 10)}.xlsx`
    link.click()
    URL.revokeObjectURL(url)
  } catch (err) {
    reportError.value = err?.data?.error?.message || t('entities.sales.reportError')
  } finally {
    generatingReport.value = false
  }
}

// color="neutral" + variant="ghost" hovers with an opaque bg-elevated,
// noticeably darker than the primary/error ghost buttons next to it
// (their hover is a 10%-opacity tint) — match that so the offer/receipt
// icons don't stand out with a dark hover circle.
const rowActionNeutralUi = { base: 'hover:bg-neutral-500/10 active:bg-neutral-500/10 focus-visible:bg-neutral-500/10' }

function onSaleCancelled() {
  table.value?.reload()
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ t(config.label) }}</h1>
      <div class="flex items-center gap-2">
        <UDropdownMenu v-if="canGenerateSalesReport" :items="reportItems">
          <UButton
            icon="i-lucide-file-down"
            color="neutral"
            variant="outline"
            :loading="generatingReport"
          >
            {{ t('entities.sales.report') }}
          </UButton>
        </UDropdownMenu>
        <UButton v-if="canCreate" icon="i-lucide-plus" to="/sales/new">{{ t('entities.sales.create') }}</UButton>
      </div>
    </div>
    <p v-if="reportError" class="text-sm text-error">{{ reportError }}</p>
    <EntityDataTable ref="table" entity="sales" :row-to="(item) => `/sales/${item.id}`">
      <template #row-actions="{ item }">
        <UButton
          icon="i-lucide-pencil"
          color="primary"
          variant="ghost"
          size="xs"
          :aria-label="t('common.edit')"
          :to="`/sales/${item.id}`"
        />
        <UButton
          v-if="canGenerateInvoice"
          icon="i-lucide-file-text"
          color="neutral"
          variant="ghost"
          size="xs"
          :ui="rowActionNeutralUi"
          :aria-label="t('entities.sales.generateInvoice')"
          @click="invoiceModal.open(item, 'offer')"
        />
        <UButton
          v-if="canGenerateInvoice && item.status !== 'SALE_STATUS_DRAFT'"
          icon="i-lucide-receipt"
          color="neutral"
          variant="ghost"
          size="xs"
          :ui="rowActionNeutralUi"
          :aria-label="t('entities.sales.generateReceipt')"
          @click="invoiceModal.open(item, 'receipt')"
        />
        <UButton
          v-if="canUpdate && !isSaleTerminal(item.status)"
          icon="i-lucide-ban"
          color="error"
          variant="ghost"
          size="xs"
          :aria-label="t('entities.sales.cancel')"
          @click="cancelModal.open(item)"
        />
      </template>
    </EntityDataTable>

    <SaleInvoiceModal ref="invoiceModal" />
    <SaleCancelModal ref="cancelModal" @cancelled="onSaleCancelled" />
  </div>
</template>
