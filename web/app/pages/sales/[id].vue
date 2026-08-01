<script setup>
// Sale detail page (TD §12.3): items are always readonly (immutable once
// created). Status changes through a client-side state machine that's
// UX only — it offers every non-current status once the sale isn't
// already terminal, but the server is the authoritative check. Cancel is
// a separate action from UpdateStatus, with a reason.
import { ENUMS } from '~/config/enums.js'
import { STATUS_COLOR_MAP } from '~/design/tokens.js'
import { toQuantity } from '~/utils/quantityAmount.js'
import { SALE_TERMINAL_STATUSES, isSaleTerminal } from '~/utils/saleStatus.js'

const route = useRoute()
const { t } = useI18n()
const { can } = usePermission()
const saleApi = useSaleApi()
const { formatMoney } = useFormatMoney()
const { formatDate } = useFormatDate()
const { handle } = useApiErrorHandler()

const loading = ref(true)
const sale = ref(null)

const canUpdate = computed(() => can('sales:update'))
const isTerminal = computed(() => sale.value && isSaleTerminal(sale.value.status))

// ENUMS.SaleStatus.values is already declared in flow order (Draft ->
// Paid -> Shipped -> Completed), with Cancelled/Refunded as terminal
// side-branches reachable from any non-terminal status — shown as a
// tooltip next to the status badge instead of a separate diagram.
const statusFlowText = computed(() => {
  const forward = ENUMS.SaleStatus.values.filter((v) => !isSaleTerminal(v)).map((v) => t(`enums.status.${v}`))
  const terminal = SALE_TERMINAL_STATUSES.map((v) => t(`enums.status.${v}`))
  return `${forward.join(' → ')}\n${t('entities.sales.statusFlowTerminalNote')}: ${terminal.join(', ')}`
})

const statusModalOpen = ref(false)
const pendingStatus = ref(null)
const statusOptions = computed(() =>
  ENUMS.SaleStatus.values.filter((v) => v !== sale.value?.status).map((v) => ({ label: t(`enums.status.${v}`), value: v })),
)
const changingStatus = ref(false)

async function load() {
  loading.value = true
  try {
    sale.value = await saleApi.get(route.params.id)
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

async function onChangeStatus() {
  if (!pendingStatus.value) return
  changingStatus.value = true
  try {
    sale.value = await saleApi.updateStatus(sale.value.id, pendingStatus.value, sale.value.etag)
    statusModalOpen.value = false
    pendingStatus.value = null
  } catch (err) {
    handle(err)
  } finally {
    changingStatus.value = false
  }
}

const cancelModal = ref(null)
function onSaleCancelled(updated) {
  sale.value = updated
}

const canGenerateInvoice = computed(() => can('invoices:generate'))
// A receipt documents a payment already received — the backend
// (InvoicesService.Generate) rejects it outright while the sale is still
// Draft, so this mirrors that one rule rather than re-deriving a wider
// notion of "paid" client-side.
const canGenerateReceipt = computed(() => canGenerateInvoice.value && sale.value?.status !== 'SALE_STATUS_DRAFT')
const invoiceModal = ref(null)

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else-if="sale">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-semibold">{{ t('entities.sales.detailTitle') }} #{{ sale.number }}</h1>
        <div class="flex items-center gap-1.5">
          <StatusBadge :status="sale.status" :map="STATUS_COLOR_MAP.sale" />
          <UTooltip :text="statusFlowText">
            <UIcon name="i-lucide-circle-help" class="size-4 text-neutral-400" />
          </UTooltip>
        </div>
      </div>

      <FormGrid>
        <div>
          <div class="text-sm text-neutral-500">{{ t('fields.client') }}</div>
          <RelationLabel :value="sale.clientId" relation="clients" />
        </div>
        <div>
          <div class="text-sm text-neutral-500">{{ t('fields.warehouse') }}</div>
          <RelationLabel :value="sale.warehouseId" relation="warehouses" />
        </div>
        <div v-if="sale.partnerId">
          <div class="text-sm text-neutral-500">{{ t('fields.partner') }}</div>
          <RelationLabel :value="sale.partnerId" relation="partners" />
        </div>
        <div>
          <div class="text-sm text-neutral-500">{{ t('fields.createdAt') }}</div>
          <span>{{ formatDate(sale.createdAt, 'long') }}</span>
        </div>
      </FormGrid>

      <div>
        <h2 class="text-sm font-medium text-neutral-600 mb-2">{{ t('entities.sales.items') }}</h2>
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-neutral-200 text-left">
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.sku') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.quantity') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.priceAmount') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.discountPercentage') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in sale.items" :key="index" class="border-b border-neutral-100">
              <td class="py-2 px-3"><RelationLabel :value="item.skuId" relation="productSkus" /></td>
              <td class="py-2 px-3">{{ toQuantity(item.quantity) }}</td>
              <td class="py-2 px-3"><MoneyAmountLabel :value="item.priceAmount" /></td>
              <td class="py-2 px-3">{{ item.discountPercentage }}%</td>
            </tr>
          </tbody>
        </table>
        <p class="text-right font-semibold mt-2">
          {{ t('fields.totalAmount') }}: <MoneyAmountLabel :value="sale.totalAmount" />
        </p>
      </div>

      <div class="flex gap-2">
        <UButton v-if="canGenerateInvoice" variant="soft" icon="i-lucide-file-text" @click="invoiceModal.open(sale, 'offer')">
          {{ t('entities.sales.generateInvoice') }}
        </UButton>
        <UButton v-if="canGenerateReceipt" variant="soft" icon="i-lucide-receipt" @click="invoiceModal.open(sale, 'receipt')">
          {{ t('entities.sales.generateReceipt') }}
        </UButton>
        <template v-if="canUpdate && !isTerminal">
          <UButton variant="soft" @click="statusModalOpen = true">{{ t('entities.sales.changeStatus') }}</UButton>
          <UButton color="error" variant="soft" @click="cancelModal.open(sale)">{{ t('entities.sales.cancel') }}</UButton>
        </template>
      </div>
    </template>

    <UModal v-model:open="statusModalOpen">
      <template #content>
        <div class="p-6 space-y-4">
          <h3 class="text-lg font-semibold">{{ t('entities.sales.changeStatus') }}</h3>
          <UFormField :label="t('fields.status')">
            <USelect v-model="pendingStatus" :items="statusOptions" class="w-full" />
          </UFormField>
          <div class="flex justify-end gap-2">
            <UButton color="neutral" variant="soft" @click="statusModalOpen = false">{{ t('common.cancel') }}</UButton>
            <UButton :disabled="!pendingStatus" :loading="changingStatus" @click="onChangeStatus">
              {{ t('common.confirm') }}
            </UButton>
          </div>
        </div>
      </template>
    </UModal>

    <SaleCancelModal ref="cancelModal" @cancelled="onSaleCancelled" />
    <SaleInvoiceModal ref="invoiceModal" />
  </div>
</template>
