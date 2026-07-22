<script setup>
// Single-step Sale creation (TD §12.3): client/partner + line items all on
// one page. The client is found by email — search on blur; a match
// prefills phone/name/address read-only, no match switches those to
// editable inputs for a brand-new client. Either way there's no separate
// "create a client first" step: Create sends either an existing clientId
// or the typed newClient fields, and SalesService.Create finds-or-creates
// by email server-side. priceAmount is never sent to Create — the server
// always uses the product's current price; the preview here is for UX
// only, the authoritative totalAmount comes back on success.
//
// Each line item picks Product -> Variant -> SKU in a lockstep chain (each
// select disabled/empty until its parent is chosen — RelationSelect's
// `filter` prop scopes Variant by productId and SKU by variantId, same
// pattern used in ProductVariantsPanel.vue). Warehouse fulfillment is
// per-item and has two modes, chosen per line (TD §12.3 update: warehouse
// moved from sale-level to item-level so lines can be split across
// warehouses):
//  - auto: the seller enters one total quantity; on submit this page reads
//    current stock per warehouse for the SKU (inventoryApi.list) and
//    greedily fills from the warehouses with the most stock first until
//    the quantity is covered, producing one SaleCreateItem per warehouse
//    used. Simplest path for the common case of "just sell N of these".
//  - manual: the seller adds one or more warehouse+quantity rows
//    themselves ("склад — количество" pairs) — full control when a
//    specific warehouse must be drawn from (e.g. fulfilling from a
//    particular branch).
// Either way, what actually reaches SaleCreate.Items is always a flat
// list of {skuId, warehouseId, quantity, discountPercentage} — the
// auto/manual split is purely a web-side convenience over that shape.
import { resolveAutoAllocation } from '~/utils/autoAllocation.js'

const router = useRouter()
const { t } = useI18n()
const saleApi = useSaleApi()
const clientApi = useEntityApi('clients')
const partnerApi = useEntityApi('partners')
const inventoryApi = useEntityApi('inventory')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const partnerId = ref(null)
const selectedPartner = ref(null)
const saving = ref(false)

const clientEmail = ref('')
const clientSearching = ref(false)
const clientSearched = ref(false)
const foundClient = ref(null)
const newClientName = ref('')
const newClientPhone = ref('')
const newClientAddress = ref('')

// A client section is only required once the seller starts typing an
// email — leaving it blank means "sell directly to the partner"
// (partnerId set, no clientId/newClient on Create), one of the three
// valid Sale shapes alongside client-only and client-with-partner.
const clientProvided = computed(() => !!clientEmail.value)
// Partner-only sale: their Partner.discountPercentage is applied to every
// item automatically server-side, overriding whatever's typed below — see
// sale.Service.buildItems. Kept in sync here purely for the line-total
// preview and to grey out the now-meaningless manual input.
const partnerDiscountActive = computed(() => !clientProvided.value && !!selectedPartner.value)

watch(partnerId, async (id) => {
  if (!id) {
    selectedPartner.value = null
    return
  }
  try {
    selectedPartner.value = await partnerApi.get(id)
  } catch {
    selectedPartner.value = null
  }
})

async function searchClient() {
  clientSearched.value = false
  foundClient.value = null
  if (!clientEmail.value) return

  clientSearching.value = true
  try {
    const res = await clientApi.list({ filter: { email: clientEmail.value }, pagination: { limit: 1 } })
    foundClient.value = res.items[0] || null
  } finally {
    clientSearching.value = false
    clientSearched.value = true
  }
}

function onEmailChange() {
  // Any edit to the email invalidates the previous search result — the
  // next blur (or the submit-time fallback below) re-searches it.
  clientSearched.value = false
  foundClient.value = null
}

const clientValid = computed(() => {
  if (!clientProvided.value) return true
  if (foundClient.value) return true
  return !!(newClientName.value && newClientPhone.value && newClientAddress.value)
})

let nextItemId = 0
let nextAllocationId = 0
function blankAllocation() {
  // available is this row's on-hand quantity for the currently picked
  // warehouse (reported by WarehouseStockSelect) — null until a warehouse
  // is chosen, used to cap the sibling quantity input at what's actually
  // on the shelf there.
  return { id: nextAllocationId++, warehouseId: null, quantity: 1, available: null }
}
function blankItem() {
  return {
    id: nextItemId++,
    productId: null,
    variantId: null,
    skuId: null,
    warehouseMode: 'auto',
    // auto mode: one total quantity, warehouses picked on submit.
    quantity: 1,
    // manual mode: explicit warehouse/quantity rows.
    allocations: [blankAllocation()],
    discountPercentage: 0,
    priceAmount: null,
    currency: null,
    // Per-warehouse stock for the currently picked skuId (from
    // SkuStockSelect's `update:stock`) — [] until a SKU is chosen. Caps
    // the auto-mode quantity input and is shown as a hint for manual
    // allocation.
    stock: [],
    totalStock: null,
  }
}
const items = ref([blankItem()])

const warehouseModeOptions = computed(() => [
  { value: 'auto', label: t('entities.sales.warehouseModeAuto') },
  { value: 'manual', label: t('entities.sales.warehouseModeManual') },
])

const itemsValid = computed(
  () =>
    items.value.length > 0 &&
    items.value.every((item) => {
      if (!item.skuId) return false
      if (item.warehouseMode === 'auto') return item.quantity > 0
      return item.allocations.length > 0 && item.allocations.every((a) => a.warehouseId && a.quantity > 0)
    }),
)
const formValid = computed(
  () => itemsValid.value && clientValid.value && (clientProvided.value || !!partnerId.value),
)

function addItem() {
  items.value = [...items.value, blankItem()]
}
function removeItem(id) {
  items.value = items.value.filter((item) => item.id !== id)
}

function addAllocation(item) {
  item.allocations = [...item.allocations, blankAllocation()]
}
function removeAllocation(item, allocationId) {
  item.allocations = item.allocations.filter((a) => a.id !== allocationId)
}

function onProductSelected(item, productId) {
  item.productId = productId
  item.variantId = null
  item.skuId = null
  item.priceAmount = null
  item.currency = null
  item.stock = []
  item.totalStock = null
}

function onVariantSelected(item, variantId) {
  item.variantId = variantId
  item.skuId = null
  item.priceAmount = null
  item.currency = null
  item.stock = []
  item.totalStock = null
}

async function onSkuSelected(item, skuId) {
  item.skuId = skuId
  item.priceAmount = null
  item.currency = null
  if (!item.skuId) return
  try {
    const price = await priceApi.get(item.skuId)
    item.priceAmount = price.priceAmount
    item.currency = price.currency
  } catch {
    // No price set for this SKU yet — preview just stays blank;
    // Create will still fail server-side if it truly requires one.
  }
}

// SkuStockSelect already looked up Inventory for this SKU (to filter the
// dropdown to in-stock SKUs in the first place) — reuse that instead of a
// second fetch. Clamp whatever quantity/allocations were already typed
// down to the new total, since switching SKUs can only shrink what's
// available.
function onStockUpdated(item, { stock, totalStock }) {
  item.stock = stock
  item.totalStock = totalStock
  if (totalStock != null && item.quantity > totalStock) item.quantity = totalStock || 1
}

function itemQuantity(item) {
  if (item.warehouseMode === 'auto') return item.quantity
  return item.allocations.reduce((sum, a) => sum + (a.quantity || 0), 0)
}

function effectiveDiscount(item) {
  return partnerDiscountActive.value ? selectedPartner.value.discountPercentage : item.discountPercentage
}

function lineTotal(item) {
  if (item.priceAmount == null) return null
  return Math.round(item.priceAmount * itemQuantity(item) * (100 - effectiveDiscount(item)) / 100)
}

const { formatMoney } = useFormatMoney()

// Grand total across every line, grouped by currency (uniform in
// practice, but items can in principle price out in different
// currencies) — null lines (no price yet) are excluded rather than
// forcing the whole total to null.
const grandTotals = computed(() => {
  const totals = new Map()
  for (const item of items.value) {
    const total = lineTotal(item)
    if (total == null) continue
    totals.set(item.currency, (totals.get(item.currency) || 0) + total)
  }
  return [...totals.entries()].map(([currency, amount]) => ({ currency, amount }))
})

// Splits item.quantity across warehouses for one auto-mode line, using the
// stock already fetched by SkuStockSelect (re-fetched here instead of
// trusting the cached copy, since time may have passed between picking the
// SKU and submitting). Throws (caught by onSubmit) if total stock across
// all warehouses falls short — surfaced as a plain validation error
// instead of a confusing per-warehouse insufficient-stock error from
// Create.
async function resolveAutoAllocations(item) {
  const res = await inventoryApi.list({ filter: { skuId: item.skuId }, pagination: { limit: 200 } })
  const { allocations, shortfall } = resolveAutoAllocation(item.quantity, res.items || [])
  if (shortfall > 0) {
    throw new Error(t('entities.sales.warehouseAutoInsufficientStock'))
  }
  return allocations
}

const autoAllocationError = ref('')

async function onSubmit() {
  // Defensive re-search: covers a submit right after typing without ever
  // blurring the email field.
  if (!clientSearched.value) await searchClient()
  if (!formValid.value) return

  autoAllocationError.value = ''
  saving.value = true
  try {
    const flatItems = []
    for (const item of items.value) {
      const allocations =
        item.warehouseMode === 'auto' ? await resolveAutoAllocations(item) : item.allocations
      for (const a of allocations) {
        flatItems.push({
          skuId: item.skuId,
          warehouseId: a.warehouseId,
          quantity: a.quantity,
          discountPercentage: effectiveDiscount(item),
        })
      }
    }

    const sale = await saleApi.create({
      ...(clientProvided.value
        ? foundClient.value
          ? { clientId: foundClient.value.id }
          : {
              newClient: {
                name: newClientName.value,
                phone: newClientPhone.value,
                email: clientEmail.value,
                address: newClientAddress.value,
              },
            }
        : {}),
      partnerId: partnerId.value,
      items: flatItems,
    })
    router.push(`/sales/${sale.id}`)
  } catch (err) {
    // resolveAutoAllocations throws a plain Error (no server round-trip);
    // everything else is a $fetch failure from saleApi.create, routed
    // through the shared handler like every other API error on this page.
    if (err?.data?.error) {
      handle(err)
    } else {
      autoAllocationError.value = err?.message || t('errors.generic')
    }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <h1 class="text-xl font-semibold">{{ t('entities.sales.create') }}</h1>

    <div class="space-y-4">
      <h2 class="font-medium">{{ t('fields.client') }}</h2>
      <p class="text-sm text-neutral-500">{{ t('entities.sales.clientOptionalHint') }}</p>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField :label="t('fields.email')">
          <UInput
            v-model="clientEmail"
            type="email"
            icon="i-lucide-search"
            :loading="clientSearching"
            class="w-full"
            @update:model-value="onEmailChange"
            @blur="searchClient"
          />
        </UFormField>

        <template v-if="clientSearched && foundClient">
          <UFormField :label="t('fields.phone')"><UInput :model-value="foundClient.phone" readonly class="w-full" /></UFormField>
          <UFormField :label="t('fields.name')"><UInput :model-value="foundClient.name" readonly class="w-full" /></UFormField>
          <UFormField :label="t('fields.address')"><UInput :model-value="foundClient.address" readonly class="w-full" /></UFormField>
        </template>
        <template v-else-if="clientSearched">
          <UFormField :label="t('fields.phone')" required>
            <UInput v-model="newClientPhone" class="w-full" />
          </UFormField>
          <UFormField :label="t('fields.name')" required>
            <UInput v-model="newClientName" class="w-full" />
          </UFormField>
          <UFormField :label="t('fields.address')" required>
            <UInput v-model="newClientAddress" class="w-full" />
          </UFormField>
        </template>
        <template v-else>
          <UFormField :label="t('fields.phone')"><UInput disabled class="w-full" /></UFormField>
          <UFormField :label="t('fields.name')"><UInput disabled class="w-full" /></UFormField>
          <UFormField :label="t('fields.address')"><UInput disabled class="w-full" /></UFormField>
        </template>
      </div>
      <p v-if="clientSearched && !foundClient" class="text-sm text-neutral-500">{{ t('entities.sales.newClient') }}</p>
    </div>

    <div class="space-y-4">
      <h2 class="font-medium">{{ t('fields.partner') }}</h2>
      <FormGrid>
        <RelationSelect v-model="partnerId" relation="partners" :label="t('fields.partner')" searchable />
      </FormGrid>
      <p v-if="clientProvided && selectedPartner" class="text-sm text-neutral-500">
        {{ t('entities.sales.partnerCommissionHint', { rate: selectedPartner.commissionPercentage }) }}
      </p>
      <p v-else-if="partnerDiscountActive" class="text-sm text-neutral-500">
        {{ t('entities.sales.partnerDiscountHint', { rate: selectedPartner.discountPercentage }) }}
      </p>
    </div>

    <div class="space-y-4">
      <h2 class="font-medium">{{ t('entities.sales.items') }}</h2>
      <div v-for="(item, index) in items" :key="item.id" class="rounded-md border border-default p-3 space-y-3">
        <div class="flex justify-between items-center">
          <span class="text-sm font-medium text-neutral-500">{{ t('entities.sales.items') }} #{{ index + 1 }}</span>
          <UButton icon="i-lucide-x" color="error" variant="ghost" @click="removeItem(item.id)" />
        </div>
        <FormGrid>
          <RelationSelect
            :model-value="item.productId"
            relation="products"
            :label="t('fields.product')"
            required
            searchable
            @update:model-value="(v) => onProductSelected(item, v)"
          />
          <RelationSelect
            :model-value="item.variantId"
            relation="productVariants"
            :label="t('fields.variant')"
            :filter="{ productIds: [item.productId] }"
            :disabled="!item.productId"
            required
            @update:model-value="(v) => onVariantSelected(item, v)"
          />
        </FormGrid>
        <FormGrid>
          <SkuStockSelect
            :model-value="item.skuId"
            :variant-id="item.variantId"
            :label="t('fields.sku')"
            required
            @update:model-value="(v) => onSkuSelected(item, v)"
            @update:stock="(v) => onStockUpdated(item, v)"
          />
          <UFormField
            :label="t('fields.discountPercentage')"
            :hint="partnerDiscountActive ? t('entities.sales.discountFromPartnerHint') : undefined"
          >
            <UInputNumber
              :model-value="partnerDiscountActive ? selectedPartner.discountPercentage : item.discountPercentage"
              class="w-full"
              :min="0"
              :max="100"
              :disabled="partnerDiscountActive"
              @update:model-value="(v) => (item.discountPercentage = v)"
            />
          </UFormField>
        </FormGrid>

        <div class="space-y-2">
          <UFormField :label="t('entities.sales.warehouseMode')">
            <URadioGroup v-model="item.warehouseMode" orientation="horizontal" :items="warehouseModeOptions" />
          </UFormField>

          <UFormField
            v-if="item.warehouseMode === 'auto'"
            :label="t('fields.quantity')"
            :hint="item.totalStock != null ? t('entities.sales.inStock', { count: item.totalStock }) : undefined"
          >
            <UInputNumber
              v-model="item.quantity"
              class="w-full max-w-40"
              :min="1"
              :max="item.totalStock ?? undefined"
              :disabled="!item.skuId || !item.totalStock"
            />
          </UFormField>

          <div v-else class="space-y-2">
            <div v-for="allocation in item.allocations" :key="allocation.id" class="flex items-end gap-2">
              <WarehouseStockSelect
                v-model="allocation.warehouseId"
                :sku-id="item.skuId"
                :label="t('fields.warehouse')"
                required
                only-in-stock
                class="flex-1"
                @update:available="(v) => (allocation.available = v)"
              />
              <UFormField
                :label="t('fields.quantity')"
                class="w-32"
                :hint="allocation.available != null ? t('entities.sales.inStock', { count: allocation.available }) : undefined"
              >
                <UInputNumber
                  v-model="allocation.quantity"
                  class="w-full"
                  :min="1"
                  :max="allocation.available ?? undefined"
                  :disabled="!allocation.warehouseId"
                />
              </UFormField>
              <UButton
                icon="i-lucide-x"
                color="error"
                variant="ghost"
                :disabled="item.allocations.length <= 1"
                @click="removeAllocation(item, allocation.id)"
              />
            </div>
            <UButton
              variant="soft"
              size="xs"
              icon="i-lucide-plus"
              :disabled="!item.skuId"
              @click="addAllocation(item)"
            >
              {{ t('entities.sales.addWarehouse') }}
            </UButton>
          </div>

          <!-- Per-warehouse breakdown for the picked SKU — mainly useful
          while distributing a line manually across warehouses, but shown
          in both modes since it's already fetched either way. -->
          <p v-if="item.stock.length" class="text-xs text-neutral-500">
            {{ t('entities.sales.availableByWarehouse') }}:
            {{ item.stock.map((s) => `${s.warehouseLabel} — ${s.quantity}`).join(', ') }}
          </p>
        </div>

        <p v-if="item.priceAmount != null" class="text-sm text-neutral-500">
          {{ t('fields.priceAmount') }}: {{ formatMoney(item.priceAmount, item.currency) }}
          — {{ t('entities.sales.lineTotal') }}: {{ formatMoney(lineTotal(item), item.currency) }}
        </p>
      </div>
      <UButton variant="soft" icon="i-lucide-plus" @click="addItem">{{ t('entities.sales.addItem') }}</UButton>
    </div>

    <p v-if="autoAllocationError" class="text-sm text-error">{{ autoAllocationError }}</p>

    <div class="flex justify-between items-center pt-2">
      <div v-if="grandTotals.length" class="text-lg font-semibold">
        {{ t('entities.sales.grandTotal') }}:
        <span v-for="gt in grandTotals" :key="gt.currency" class="ms-1">{{ formatMoney(gt.amount, gt.currency) }}</span>
      </div>
      <div v-else />
      <UButton :disabled="!formValid" :loading="saving" @click="onSubmit">{{ t('common.create') }}</UButton>
    </div>
  </div>
</template>
