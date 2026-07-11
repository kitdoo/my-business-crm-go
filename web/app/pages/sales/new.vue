<script setup>
// 2-step Sale creation wizard (TD §12.3): step 1 picks client/warehouse/
// partner (with inline "create a new client" so the operator doesn't
// lose context), step 2 adds line items with a live price preview.
// priceAmount is never sent to Create — the server always uses the
// product's current price; the preview here is for UX only, the
// authoritative totalAmount comes back from the server on success.
const router = useRouter()
const { t } = useI18n()
const saleApi = useSaleApi()
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const step = ref(1)
const clientId = ref(null)
const warehouseId = ref(null)
const partnerId = ref(null)
const clientModalOpen = ref(false)
const saving = ref(false)

let nextItemId = 0
const items = ref([{ id: nextItemId++, productId: null, quantity: 1, discountPercentage: 0, priceAmount: null, currency: null }])

const step1Valid = computed(() => !!clientId.value && !!warehouseId.value)
const step2Valid = computed(() => items.value.length > 0 && items.value.every((item) => item.productId && item.quantity > 0))

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

async function onCreateClientSaved(client) {
  clientId.value = client.id
  clientModalOpen.value = false
}

async function onSubmit() {
  saving.value = true
  try {
    const sale = await saleApi.create({
      clientId: clientId.value,
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

    <div v-if="step === 1" class="space-y-4">
      <FormGrid>
        <RelationSelect v-model="clientId" relation="clients" :label="t('fields.client')" required />
        <RelationSelect v-model="warehouseId" relation="warehouses" :label="t('fields.warehouse')" required />
        <RelationSelect v-model="partnerId" relation="partners" :label="t('fields.partner')" />
      </FormGrid>
      <UButton variant="soft" icon="i-lucide-user-plus" @click="clientModalOpen = true">
        {{ t('entities.sales.newClient') }}
      </UButton>

      <div class="flex justify-end">
        <UButton :disabled="!step1Valid" @click="step = 2">{{ t('common.next') }}</UButton>
      </div>
    </div>

    <div v-else class="space-y-4">
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

      <div class="flex items-center justify-between pt-2">
        <UButton color="neutral" variant="soft" @click="step = 1">{{ t('common.back') }}</UButton>
        <UButton :disabled="!step2Valid" :loading="saving" @click="onSubmit">{{ t('common.create') }}</UButton>
      </div>
    </div>

    <UModal v-model:open="clientModalOpen">
      <template #content>
        <div class="p-6 space-y-4">
          <h2 class="text-lg font-semibold">{{ t('entities.clients.create') }}</h2>
          <EntityForm entity="clients" mode="drawer" @saved="onCreateClientSaved" @cancel="clientModalOpen = false" />
        </div>
      </template>
    </UModal>
  </div>
</template>
