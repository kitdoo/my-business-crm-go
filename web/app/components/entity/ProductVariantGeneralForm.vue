<script setup>
// Bespoke replacement for generic <EntityForm entity="productVariants">
// (mirrors ProductGeneralForm.vue's rationale) — a variant needs
// sku(+price on create)/attributes/images grouped together, plus the
// required-price side effect on create (a variant with no price isn't
// sellable, and Price has no tab to reach before the variant exists).
// productId comes from the parent Product's page: a query param on
// create (props.productId, there's no record yet), the loaded record's
// own productId on edit (createOnly — see useEntityForm's applyRecord —
// so it never lands in the reactive `form`, only on `record`). Either
// way, never picked by hand here.
import { getEnumOptions } from '~/config/enums.js'
import { toBasisPoints } from '~/utils/priceAmount.js'

const props = defineProps({
  id: { type: String, default: null },
  productId: { type: String, default: null },
  // Used from ProductVariantsTab: create in place instead of navigating to
  // the full /product-variants/:id page, so the operator can keep adding
  // variants without leaving the product. Editing an existing variant
  // (price tab, etc.) always happens on the full page — never embedded.
  embedded: { type: Boolean, default: false },
})
const emit = defineEmits(['created', 'cancel'])

const { t } = useI18n()
const router = useRouter()
const toast = useToast()
const config = getEntityConfig('productVariants')
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()
const api = useEntityApi('productVariants')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const { form, record, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm('productVariants', props.id, { productId: props.productId })

// props.productId only exists on create (query param); on edit it comes
// from the loaded record instead (see the comment above).
const productId = computed(() => props.productId || record.value?.productId)

const priceAmount = ref(null)
const discountAmount = ref(null)
const priceError = ref('')

const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))

onMounted(() => {
  if (!isCreate) load()
})

async function onSubmit() {
  priceError.value = ''
  if (isCreate && (priceAmount.value == null || priceAmount.value < 0)) {
    priceError.value = t('forms.validationError')
    return
  }
  try {
    const record = await save()
    if (!record) return

    if (isCreate) {
      try {
        await priceApi.create(record.id, toBasisPoints(priceAmount.value), toBasisPoints(discountAmount.value))
      } catch (err) {
        handle(err)
        return
      }

      // Create has no status field server-side (ProductVariantNew always
      // starts Draft, see entities/product_variant.go) — a variant is
      // sellable as soon as it has a price, so flip it to Active right
      // away instead of leaving it stuck in Draft.
      let created = record
      try {
        created = await api.update(record.id, { status: 'PRODUCT_VARIANT_STATUS_ACTIVE' }, ['status'], record.etag)
      } catch (err) {
        handle(err)
        return
      }

      if (props.embedded) {
        emit('created', created)
        return
      }
      router.push(`${config.route}/${created.id}`)
      return
    }

    toast.add({ title: t('common.saved'), color: 'success' })
  } catch {
    // handled by useApiErrorHandler inside save()
  }
}

async function onDelete() {
  try {
    await api.remove(props.id, etag.value)
    router.push(`/products/${productId.value}`)
  } catch (err) {
    handle(err)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <form v-else class="space-y-6" @submit.prevent="onSubmit">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UFormField :label="t('fields.sku')" required :error="fieldErrors.sku">
          <UInput v-model="form.sku" class="w-full" required :maxlength="64" :readonly="!isCreate" />
        </UFormField>
        <template v-if="isCreate">
          <UFormField :label="t('fields.priceAmount')" required :error="priceError">
            <UInputNumber v-model="priceAmount" class="w-full" :min="0" :step="0.01" />
          </UFormField>
          <UFormField :label="t('fields.discountAmount')">
            <UInputNumber v-model="discountAmount" class="w-full" :min="0" :step="0.01" />
          </UFormField>
        </template>
        <UFormField v-else :label="t('fields.status')" :error="fieldErrors.status">
          <USelect
            v-model="form.status"
            :items="getEnumOptions('ProductVariantStatus').map((v) => ({ label: t(`enums.status.${v}`), value: v }))"
            class="w-full"
          />
        </UFormField>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <AttributeDetailsInput
          v-model="form.attributes"
          :locales="runtimeConfig.public.supportedLocales"
          :label="t('fields.attributes')"
        />
        <UFormField :label="t('fields.images')">
          <ProductImageUploader v-model="form.imageIds" />
        </UFormField>
      </div>

      <div class="flex items-center justify-between">
        <UButton v-if="canDelete" color="error" variant="soft" @click="confirmDeleteOpen = true">
          {{ t('common.delete') }}
        </UButton>
        <div class="ml-auto flex gap-2">
          <UButton v-if="embedded" color="neutral" variant="soft" @click="emit('cancel')">
            {{ t('common.cancel') }}
          </UButton>
          <UButton v-else color="neutral" variant="soft" :to="`/products/${productId}`">
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
