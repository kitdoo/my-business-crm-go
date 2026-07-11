<script setup>
// Price tab on the Product detail page (TD §12.1): current price
// (create-if-none, else edit) + a read-only history table underneath.
// Not a generic EntityForm — ProductPrice isn't a generic entity — but
// reuses the same buildUpdateMask/etag/useApiErrorHandler conventions.
import { buildUpdateMask } from '~/utils/buildUpdateMask.js'

const props = defineProps({
  productId: { type: String, required: true },
})

const { t } = useI18n()
const { can } = usePermission()
const { formatMoney } = useFormatMoney()
const { formatDate } = useFormatDate()
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const loading = ref(true)
const saving = ref(false)
const price = ref(null) // null = no price set yet for this product
const form = ref({ priceAmount: null, discountAmount: null })
const original = ref(null)
const fieldErrors = ref({})

const canWrite = computed(() => can('prices:create') || can('prices:update'))

const historyOpen = ref(false)
const historyLoading = ref(false)
const historyItems = ref([])

// Basis points (backend/wire) <-> plain currency amount (what the user
// types), per TD §5.4 — never send/display raw basis points to the user.
function toAmount(basisPoints) {
  return basisPoints == null ? null : basisPoints / 100
}
function toBasisPoints(amount) {
  return amount == null ? null : Math.round(amount * 100)
}

async function load() {
  loading.value = true
  try {
    const rec = await priceApi.get(props.productId)
    price.value = rec
    form.value = { priceAmount: toAmount(rec.priceAmount), discountAmount: toAmount(rec.discountAmount) }
    original.value = { ...form.value }
  } catch (err) {
    if (err?.data?.error?.code === 'NOT_FOUND') {
      price.value = null
      form.value = { priceAmount: null, discountAmount: null }
      original.value = null
    } else {
      handle(err)
    }
  } finally {
    loading.value = false
  }
}

async function onSubmit() {
  fieldErrors.value = {}
  saving.value = true
  try {
    if (!price.value) {
      const rec = await priceApi.create(
        props.productId,
        toBasisPoints(form.value.priceAmount),
        toBasisPoints(form.value.discountAmount),
      )
      price.value = rec
      form.value = { priceAmount: toAmount(rec.priceAmount), discountAmount: toAmount(rec.discountAmount) }
      original.value = { ...form.value }
      return
    }

    const mask = buildUpdateMask(original.value, form.value, ['priceAmount', 'discountAmount'])
    if (mask.length === 0) return
    const fields = {}
    for (const key of mask) fields[key] = toBasisPoints(form.value[key])

    const rec = await priceApi.update(price.value.id, fields, mask, price.value.etag)
    price.value = rec
    form.value = { priceAmount: toAmount(rec.priceAmount), discountAmount: toAmount(rec.discountAmount) }
    original.value = { ...form.value }
  } catch (err) {
    handle(err, {
      onFieldError: (field, message) => {
        fieldErrors.value = { ...fieldErrors.value, [field]: message }
      },
    })
  } finally {
    saving.value = false
  }
}

async function toggleHistory() {
  historyOpen.value = !historyOpen.value
  if (historyOpen.value && historyItems.value.length === 0) {
    historyLoading.value = true
    try {
      const res = await priceApi.history(props.productId, { pagination: { limit: 25 } })
      historyItems.value = res.items
    } catch (err) {
      handle(err)
    } finally {
      historyLoading.value = false
    }
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <p v-if="!price" class="text-sm text-neutral-500">{{ t('prices.noPriceYet') }}</p>
      <p v-else class="text-lg font-semibold">
        {{ formatMoney(price.priceAmount, price.currency) }}
        <span v-if="price.discountAmount" class="text-sm text-neutral-500 ml-2">
          {{ t('prices.discount') }}: {{ formatMoney(price.discountAmount, price.currency) }}
        </span>
      </p>

      <form v-if="canWrite" class="space-y-4" @submit.prevent="onSubmit">
        <FormGrid>
          <UFormField :label="t('fields.priceAmount')" :error="fieldErrors.priceAmount" required>
            <UInputNumber v-model="form.priceAmount" class="w-full" :min="0" :step="0.01" />
          </UFormField>
          <UFormField :label="t('fields.discountAmount')" :error="fieldErrors.discountAmount">
            <UInputNumber v-model="form.discountAmount" class="w-full" :min="0" :step="0.01" />
          </UFormField>
        </FormGrid>
        <UButton type="submit" :loading="saving">{{ t('common.save') }}</UButton>
      </form>

      <div>
        <UButton variant="link" :icon="historyOpen ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'" @click="toggleHistory">
          {{ t('prices.history') }}
        </UButton>
        <div v-if="historyOpen" class="mt-2">
          <div v-if="historyLoading" class="py-4 text-center text-neutral-500">{{ t('common.loading') }}</div>
          <div v-else-if="historyItems.length === 0" class="py-4 text-center text-neutral-500">{{ t('common.empty') }}</div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr class="border-b border-neutral-200 text-left">
                <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.priceAmount') }}</th>
                <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.discountAmount') }}</th>
                <th class="py-2 px-3 font-medium text-neutral-600">{{ t('fields.createdAt') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in historyItems" :key="item.id" class="border-b border-neutral-100">
                <td class="py-2 px-3">{{ formatMoney(item.priceAmount, item.currency) }}</td>
                <td class="py-2 px-3">{{ item.discountAmount ? formatMoney(item.discountAmount, item.currency) : '—' }}</td>
                <td class="py-2 px-3">{{ formatDate(item.createdAt, 'long') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
