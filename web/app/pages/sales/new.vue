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
const router = useRouter()
const { t } = useI18n()
const saleApi = useSaleApi()
const clientApi = useEntityApi('clients')
const inventoryApi = useEntityApi('inventory')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const partnerId = ref(null)
const saving = ref(false)

const clientEmail = ref('')
const clientSearching = ref(false)
const clientSearched = ref(false)
const foundClient = ref(null)
const newClientName = ref('')
const newClientPhone = ref('')
const newClientAddress = ref('')

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
  if (foundClient.value) return true
  return !!(clientEmail.value && newClientName.value && newClientPhone.value && newClientAddress.value)
})

let nextItemId = 0
let nextAllocationId = 0
function blankAllocation() {
  return { id: nextAllocationId++, warehouseId: null, quantity: 1 }
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
const formValid = computed(() => clientValid.value && itemsValid.value)

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
}

function onVariantSelected(item, variantId) {
  item.variantId = variantId
  item.skuId = null
  item.priceAmount = null
  item.currency = null
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

function itemQuantity(item) {
  if (item.warehouseMode === 'auto') return item.quantity
  return item.allocations.reduce((sum, a) => sum + (a.quantity || 0), 0)
}

function lineTotal(item) {
  if (item.priceAmount == null) return null
  return Math.round(item.priceAmount * itemQuantity(item) * (100 - item.discountPercentage) / 100)
}

const { formatMoney } = useFormatMoney()

// Splits item.quantity across warehouses for one auto-mode line: reads
// current stock per warehouse for the SKU and fills from the
// best-stocked warehouse down until the requested quantity is covered.
// Throws (caught by onSubmit) if total stock across all warehouses falls
// short — surfaced as a plain validation error instead of a confusing
// per-warehouse insufficient-stock error from Create.
async function resolveAutoAllocations(item) {
  const res = await inventoryApi.list({ filter: { skuId: item.skuId }, pagination: { limit: 200 } })
  const stock = (res.items || [])
    .filter((s) => s.quantity > 0)
    .sort((a, b) => b.quantity - a.quantity)

  let remaining = item.quantity
  const allocations = []
  for (const s of stock) {
    if (remaining <= 0) break
    const take = Math.min(remaining, s.quantity)
    allocations.push({ warehouseId: s.warehouseId, quantity: take })
    remaining -= take
  }
  if (remaining > 0) {
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
          discountPercentage: item.discountPercentage,
        })
      }
    }

    const sale = await saleApi.create({
      ...(foundClient.value
        ? { clientId: foundClient.value.id }
        : {
            newClient: {
              name: newClientName.value,
              phone: newClientPhone.value,
              email: clientEmail.value,
              address: newClientAddress.value,
            },
          }),
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
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField :label="t('fields.email')" required>
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
        <RelationSelect v-model="partnerId" relation="partners" :label="t('fields.partner')" />
      </FormGrid>
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
          <RelationSelect
            :model-value="item.skuId"
            relation="productSkus"
            :label="t('fields.sku')"
            :filter="{ variantIds: [item.variantId] }"
            :disabled="!item.variantId"
            required
            @update:model-value="(v) => onSkuSelected(item, v)"
          />
          <UFormField :label="t('fields.discountPercentage')">
            <UInputNumber v-model="item.discountPercentage" class="w-full" :min="0" :max="100" />
          </UFormField>
        </FormGrid>

        <div class="space-y-2">
          <UFormField :label="t('entities.sales.warehouseMode')">
            <URadioGroup v-model="item.warehouseMode" orientation="horizontal" :items="warehouseModeOptions" />
          </UFormField>

          <UFormField v-if="item.warehouseMode === 'auto'" :label="t('fields.quantity')">
            <UInputNumber v-model="item.quantity" class="w-full max-w-40" :min="1" :disabled="!item.skuId" />
          </UFormField>

          <div v-else class="space-y-2">
            <div v-for="allocation in item.allocations" :key="allocation.id" class="flex items-end gap-2">
              <RelationSelect
                v-model="allocation.warehouseId"
                relation="warehouses"
                :label="t('fields.warehouse')"
                :disabled="!item.skuId"
                required
                class="flex-1"
              />
              <UFormField :label="t('fields.quantity')" class="w-32">
                <UInputNumber v-model="allocation.quantity" class="w-full" :min="1" :disabled="!item.skuId" />
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
        </div>

        <p v-if="item.priceAmount != null" class="text-sm text-neutral-500">
          {{ t('fields.priceAmount') }}: {{ formatMoney(item.priceAmount, item.currency) }}
          — {{ t('entities.sales.lineTotal') }}: {{ formatMoney(lineTotal(item), item.currency) }}
        </p>
      </div>
      <UButton variant="soft" icon="i-lucide-plus" @click="addItem">{{ t('entities.sales.addItem') }}</UButton>
    </div>

    <p v-if="autoAllocationError" class="text-sm text-error">{{ autoAllocationError }}</p>

    <div class="flex justify-end pt-2">
      <UButton :disabled="!formValid" :loading="saving" @click="onSubmit">{{ t('common.create') }}</UButton>
    </div>
  </div>
</template>
