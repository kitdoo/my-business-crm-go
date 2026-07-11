<script setup>
// Sale detail page (TD §12.3): items are always readonly (immutable once
// created). Status changes through a client-side state machine that's
// UX only — it offers every non-current status once the sale isn't
// already terminal, but the server is the authoritative check. Cancel is
// a separate action from UpdateStatus, with a reason.
import { ENUMS } from '~/config/enums.js'
import { STATUS_COLOR_MAP } from '~/design/tokens.js'

const route = useRoute()
const { t } = useI18n()
const { can } = usePermission()
const saleApi = useSaleApi()
const { formatMoney } = useFormatMoney()
const { formatDate } = useFormatDate()
const { handle } = useApiErrorHandler()

const loading = ref(true)
const sale = ref(null)

const TERMINAL_STATUSES = ['SALE_STATUS_CANCELLED', 'SALE_STATUS_REFUNDED']
const canUpdate = computed(() => can('sales:update'))
const isTerminal = computed(() => sale.value && TERMINAL_STATUSES.includes(sale.value.status))

const statusModalOpen = ref(false)
const pendingStatus = ref(null)
const statusOptions = computed(() =>
  ENUMS.SaleStatus.values.filter((v) => v !== sale.value?.status).map((v) => ({ label: t(`enums.status.${v}`), value: v })),
)
const changingStatus = ref(false)

const cancelModalOpen = ref(false)
const cancelReason = ref('')
const cancelling = ref(false)

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

async function onCancelConfirm() {
  cancelling.value = true
  try {
    sale.value = await saleApi.cancel(sale.value.id, cancelReason.value, sale.value.etag)
    cancelModalOpen.value = false
    cancelReason.value = ''
  } catch (err) {
    handle(err)
  } finally {
    cancelling.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else-if="sale">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-semibold">{{ t('entities.sales.detailTitle') }}</h1>
        <StatusBadge :status="sale.status" :map="STATUS_COLOR_MAP.sale" />
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
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.product') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.quantity') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.priceAmount') }}</th>
              <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.discountPercentage') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, index) in sale.items" :key="index" class="border-b border-neutral-100">
              <td class="py-2 px-3"><RelationLabel :value="item.productId" relation="products" /></td>
              <td class="py-2 px-3">{{ item.quantity }}</td>
              <td class="py-2 px-3"><MoneyAmountLabel :value="item.priceAmount" /></td>
              <td class="py-2 px-3">{{ item.discountPercentage }}%</td>
            </tr>
          </tbody>
        </table>
        <p class="text-right font-semibold mt-2">
          {{ t('fields.totalAmount') }}: <MoneyAmountLabel :value="sale.totalAmount" />
        </p>
      </div>

      <div v-if="canUpdate && !isTerminal" class="flex gap-2">
        <UButton variant="soft" @click="statusModalOpen = true">{{ t('entities.sales.changeStatus') }}</UButton>
        <UButton color="error" variant="soft" @click="cancelModalOpen = true">{{ t('entities.sales.cancel') }}</UButton>
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

    <UModal v-model:open="cancelModalOpen">
      <template #content>
        <div class="p-6 space-y-4">
          <h3 class="text-lg font-semibold">{{ t('entities.sales.cancelConfirmTitle') }}</h3>
          <p class="text-sm text-neutral-600">{{ t('entities.sales.cancelConfirmBody') }}</p>
          <UFormField :label="t('entities.sales.cancelReason')">
            <UTextarea v-model="cancelReason" class="w-full" />
          </UFormField>
          <div class="flex justify-end gap-2">
            <UButton color="neutral" variant="soft" @click="cancelModalOpen = false">{{ t('common.back') }}</UButton>
            <UButton color="error" :loading="cancelling" @click="onCancelConfirm">{{ t('entities.sales.cancel') }}</UButton>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>
