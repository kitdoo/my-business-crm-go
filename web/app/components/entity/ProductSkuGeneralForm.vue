<script setup>
// Bespoke replacement for generic <EntityForm entity="productSkus">
// (mirrors ProductGeneralForm.vue's rationale) — one compact form for sku
// code, price, and attributes together, so there's a single lock/edit
// state and a single Save instead of the sku fields and the price sitting
// in separate always-editable widgets. No Status field: a SKU is always
// Active the moment it's saved (flipped right after create, alongside its
// required price) — the only other lifecycle action is Delete, never a
// manual draft/inactive toggle. Stock shows as a plain read-only total
// (see loadStock below) — recording movements happens on the dedicated
// /inventory-movements page, not here.
// Always rendered inline on the Product page (see ProductSkusPanel.vue) —
// there is no standalone /product-skus/:id page. variantId comes from the
// parent Variant's block: a prop on create (there's no record yet), the
// loaded record's own variantId on edit (createOnly — see useEntityForm's
// applyRecord — so it never lands in the reactive `form`, only on
// `record`). Either way, never picked by hand here. This component never
// navigates: every outcome (created/cancelled/deleted) is emitted for
// ProductSkusPanel to fold into its own list state in place.
import { buildUpdateMask } from '~/utils/buildUpdateMask.js'
import { toAmount, toBasisPoints } from '~/utils/priceAmount.js'
import { fetchTotalStock } from '~/utils/inventoryStock.js'

const props = defineProps({
  id: { type: String, default: null },
  variantId: { type: String, default: null },
})
const emit = defineEmits(['created', 'cancel', 'deleted'])

const { t } = useI18n()
const toast = useToast()
const config = getEntityConfig('productSkus')
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()
const api = useEntityApi('productSkus')
const inventoryApi = useEntityApi('inventory')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const { form, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm('productSkus', props.id, { variantId: props.variantId })

const price = ref(null) // loaded ProductPrice record (edit mode only)
const priceForm = ref({ priceAmount: null, discountAmount: null })
const priceOriginal = ref(null)
const priceFieldErrors = ref({})
const priceError = ref('') // create-mode validation only, see onSubmit

const totalStock = ref(0)

// Same locked/edit pattern as ProductGeneralForm.vue — an existing SKU
// defaults to read-only with a pencil button to unlock. Never applies to
// create: there's nothing to look at yet.
const locked = ref(!isCreate)
const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))
const canUpdate = computed(() => !isCreate && can(config.permissions.update))

async function loadPrice() {
  try {
    const rec = await priceApi.get(props.id)
    price.value = rec
    priceForm.value = { priceAmount: toAmount(rec.priceAmount), discountAmount: toAmount(rec.discountAmount) }
    priceOriginal.value = { ...priceForm.value }
  } catch (err) {
    // Every SKU gets a price at create time (see onSubmit below); a
    // missing one only happens on legacy/imported data, not through this
    // form — leave the fields empty rather than surfacing a toast for it.
    if (err?.data?.error?.code !== 'NOT_FOUND') handle(err)
  }
}

async function loadStock() {
  totalStock.value = await fetchTotalStock(inventoryApi, props.id)
}

onMounted(() => {
  if (!isCreate) {
    load()
    loadPrice()
    loadStock()
  }
})

function onEdit() {
  locked.value = false
}

function onCancelEdit() {
  fieldErrors.value = {}
  priceFieldErrors.value = {}
  load()
  if (priceOriginal.value) priceForm.value = { ...priceOriginal.value }
  locked.value = true
}

async function onSubmit() {
  priceError.value = ''
  if (isCreate && (priceForm.value.priceAmount == null || priceForm.value.priceAmount < 0)) {
    priceError.value = t('forms.validationError')
    return
  }
  try {
    const record = await save()
    if (!record) return

    if (isCreate) {
      try {
        await priceApi.create(
          record.id,
          toBasisPoints(priceForm.value.priceAmount),
          toBasisPoints(priceForm.value.discountAmount),
        )
      } catch (err) {
        handle(err)
        return
      }

      // Create has no status field server-side (ProductSKUNew always
      // starts Draft, see entities/product_sku.go) — a SKU is sellable
      // as soon as it has a price, so flip it to Active right away
      // instead of leaving it stuck in Draft.
      let created = record
      try {
        created = await api.update(record.id, { status: 'PRODUCT_SKU_STATUS_ACTIVE' }, ['status'], record.etag)
      } catch (err) {
        handle(err)
        return
      }

      emit('created', created)
      return
    }

    if (price.value) {
      priceFieldErrors.value = {}
      const mask = buildUpdateMask(priceOriginal.value, priceForm.value, ['priceAmount', 'discountAmount'])
      if (mask.length > 0) {
        const fields = {}
        for (const key of mask) fields[key] = toBasisPoints(priceForm.value[key])
        try {
          const updated = await priceApi.update(price.value.id, fields, mask, price.value.etag)
          price.value = updated
          priceForm.value = { priceAmount: toAmount(updated.priceAmount), discountAmount: toAmount(updated.discountAmount) }
          priceOriginal.value = { ...priceForm.value }
        } catch (err) {
          handle(err, {
            onFieldError: (field, message) => {
              priceFieldErrors.value = { ...priceFieldErrors.value, [field]: message }
            },
          })
          return
        }
      }
    }

    locked.value = true
    toast.add({ title: t('common.saved'), color: 'success' })
  } catch {
    // handled by useApiErrorHandler inside save()
  }
}

async function onDelete() {
  try {
    await api.remove(props.id, etag.value)
    emit('deleted')
  } catch (err) {
    handle(err)
  }
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <form v-else class="space-y-4" @submit.prevent="onSubmit">
      <div v-if="locked && canUpdate" class="flex items-center justify-end">
        <UButton icon="i-lucide-pencil" color="neutral" variant="soft" @click="onEdit">
          {{ t('common.edit') }}
        </UButton>
      </div>

      <fieldset :disabled="locked" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <UFormField :label="t('fields.sku')" required :error="fieldErrors.sku">
            <UInput v-model="form.sku" class="w-full" required :maxlength="64" :readonly="!isCreate" />
          </UFormField>
          <UFormField :label="t('fields.priceAmount')" required :error="priceError || priceFieldErrors.priceAmount">
            <UInputNumber v-model="priceForm.priceAmount" class="w-full" :min="0" :step="0.01" :disabled="locked" />
          </UFormField>
          <UFormField :label="t('fields.discountAmount')" :error="priceFieldErrors.discountAmount">
            <UInputNumber v-model="priceForm.discountAmount" class="w-full" :min="0" :step="0.01" :disabled="locked" />
          </UFormField>
        </div>

        <AttributeDetailsInput
          v-model="form.attributes"
          :locales="runtimeConfig.public.supportedLocales"
          :label="t('fields.attributes')"
          :disabled="locked"
        />
      </fieldset>

      <p v-if="!isCreate" class="text-sm text-neutral-500">
        {{ t('entities.inventory.label') }}: <span class="font-medium text-default">{{ totalStock }}</span>
      </p>

      <div v-if="!locked" class="flex items-center justify-between">
        <UButton v-if="canDelete" color="error" variant="soft" @click="confirmDeleteOpen = true">
          {{ t('common.delete') }}
        </UButton>
        <div class="ml-auto flex gap-2">
          <UButton v-if="isCreate" color="neutral" variant="soft" @click="emit('cancel')">
            {{ t('common.cancel') }}
          </UButton>
          <UButton v-else color="neutral" variant="soft" @click="onCancelEdit">
            {{ t('common.cancel') }}
          </UButton>
          <UButton type="submit" :loading="saving">{{ t('common.save') }}</UButton>
        </div>
      </div>
    </form>

    <ConfirmDialog
      v-model:open="confirmDeleteOpen"
      :title="t('common.deleteConfirmTitle')"
      :description="t('common.deleteConfirmBody')"
      @confirm="onDelete"
    />
    <EtagConflictDialog v-model:open="etagConflict" @reload="reloadAfterConflict" />
  </div>
</template>
