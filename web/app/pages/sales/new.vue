<script setup>
// Single-step Sale creation (TD §12.3): client/warehouse/partner + line
// items all on one page. The client is found by email — search on blur;
// a match prefills name/phone/address read-only, no match switches those
// to editable inputs for a brand-new client. Either way there's no
// separate "create a client first" step: Create sends either an existing
// clientId or the typed newClient fields, and SalesService.Create finds-
// or-creates by email server-side. priceAmount is never sent to Create —
// the server always uses the product's current price; the preview here
// is for UX only, the authoritative totalAmount comes back on success.
const router = useRouter()
const { t } = useI18n()
const saleApi = useSaleApi()
const clientApi = useEntityApi('clients')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const warehouseId = ref(null)
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
const items = ref([{ id: nextItemId++, productId: null, quantity: 1, discountPercentage: 0, priceAmount: null, currency: null }])

const itemsValid = computed(() => items.value.length > 0 && items.value.every((item) => item.productId && item.quantity > 0))
const formValid = computed(() => clientValid.value && !!warehouseId.value && itemsValid.value)

function addItem() {
  items.value = [...items.value, { id: nextItemId++, productId: null, quantity: 1, discountPercentage: 0, priceAmount: null, currency: null }]
}
function removeItem(id) {
  items.value = items.value.filter((item) => item.id !== id)
}

async function onProductSelected(item) {
  item.priceAmount = null
  item.currency = null
  if (!item.productId) return
  try {
    const price = await priceApi.get(item.productId)
    item.priceAmount = price.priceAmount
    item.currency = price.currency
  } catch {
    // No price set for this product yet — preview just stays blank;
    // Create will still fail server-side if it truly requires one.
  }
}

function lineTotal(item) {
  if (item.priceAmount == null) return null
  return Math.round(item.priceAmount * item.quantity * (100 - item.discountPercentage) / 100)
}

const { formatMoney } = useFormatMoney()

async function onSubmit() {
  // Defensive re-search: covers a submit right after typing without ever
  // blurring the email field.
  if (!clientSearched.value) await searchClient()
  if (!formValid.value) return

  saving.value = true
  try {
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
      warehouseId: warehouseId.value,
      partnerId: partnerId.value,
      items: items.value.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
        discountPercentage: item.discountPercentage,
      })),
    })
    router.push(`/sales/${sale.id}`)
  } catch (err) {
    handle(err)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h1 class="text-xl font-semibold">{{ t('entities.sales.create') }}</h1>

    <div class="space-y-4">
      <h2 class="font-medium">{{ t('fields.client') }}</h2>
      <FormGrid>
        <UFormField :label="t('fields.email')" required>
          <UInput v-model="clientEmail" type="email" class="w-full" @update:model-value="onEmailChange" @blur="searchClient" />
        </UFormField>
      </FormGrid>

      <p v-if="clientSearching" class="text-sm text-neutral-500">{{ t('common.loading') }}</p>
      <template v-else-if="clientSearched">
        <FormGrid v-if="foundClient">
          <UFormField :label="t('fields.name')"><UInput :model-value="foundClient.name" readonly class="w-full" /></UFormField>
          <UFormField :label="t('fields.phone')"><UInput :model-value="foundClient.phone" readonly class="w-full" /></UFormField>
          <UFormField :label="t('fields.address')"><UInput :model-value="foundClient.address" readonly class="w-full" /></UFormField>
        </FormGrid>
        <template v-else>
          <p class="text-sm text-neutral-500">{{ t('entities.sales.newClient') }}</p>
          <FormGrid>
            <UFormField :label="t('fields.name')" required>
              <UInput v-model="newClientName" class="w-full" />
            </UFormField>
            <UFormField :label="t('fields.phone')" required>
              <UInput v-model="newClientPhone" class="w-full" />
            </UFormField>
            <UFormField :label="t('fields.address')" required>
              <UInput v-model="newClientAddress" class="w-full" />
            </UFormField>
          </FormGrid>
        </template>
      </template>

      <FormGrid>
        <RelationSelect v-model="warehouseId" relation="warehouses" :label="t('fields.warehouse')" required />
        <RelationSelect v-model="partnerId" relation="partners" :label="t('fields.partner')" />
      </FormGrid>
    </div>

    <div class="space-y-4">
      <h2 class="font-medium">{{ t('entities.sales.items') }}</h2>
      <div v-for="item in items" :key="item.id" class="rounded-md border border-default p-3 space-y-3">
        <div class="flex items-start gap-2">
          <RelationSelect
            :model-value="item.productId"
            relation="products"
            :label="t('fields.product')"
            required
            class="flex-1"
            @update:model-value="
              (v) => {
                item.productId = v
                onProductSelected(item)
              }
            "
          />
          <UButton icon="i-lucide-x" color="error" variant="ghost" class="mt-6" @click="removeItem(item.id)" />
        </div>
        <FormGrid>
          <UFormField :label="t('fields.quantity')">
            <UInputNumber v-model="item.quantity" class="w-full" :min="1" />
          </UFormField>
          <UFormField :label="t('fields.discountPercentage')">
            <UInputNumber v-model="item.discountPercentage" class="w-full" :min="0" :max="100" />
          </UFormField>
        </FormGrid>
        <p v-if="item.priceAmount != null" class="text-sm text-neutral-500">
          {{ t('fields.priceAmount') }}: {{ formatMoney(item.priceAmount, item.currency) }}
          — {{ t('entities.sales.lineTotal') }}: {{ formatMoney(lineTotal(item), item.currency) }}
        </p>
      </div>
      <UButton variant="soft" icon="i-lucide-plus" @click="addItem">{{ t('entities.sales.addItem') }}</UButton>
    </div>

    <div class="flex justify-end pt-2">
      <UButton :disabled="!formValid" :loading="saving" @click="onSubmit">{{ t('common.create') }}</UButton>
    </div>
  </div>
</template>
