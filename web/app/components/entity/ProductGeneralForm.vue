<script setup>
// Bespoke replacement for generic <EntityForm entity="products"> (TD
// §12.1/§12.6) — Product needs a grouped block layout (brand/categories,
// name/description, sku(+price on create), characteristics/images) that
// the generic field-by-field <FormGrid> loop can't express, plus a
// required-price side effect on create. Reuses useEntityForm as-is (same
// load/save/etag/mask logic every other entity form gets) — only the
// template layout differs.
import { getEnumOptions } from '~/config/enums.js'
import { toAmount, toBasisPoints } from '~/utils/priceAmount.js'

const props = defineProps({
  id: { type: String, default: null },
})

const { t } = useI18n()
const router = useRouter()
const toast = useToast()
const config = getEntityConfig('products')
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()
const api = useEntityApi('products')
const priceApi = usePriceApi()
const { handle } = useApiErrorHandler()

const { form, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm('products', props.id, {})

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
      router.push(`${config.route}/${record.id}`)
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
    router.push(config.route)
  } catch (err) {
    handle(err)
  }
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <form v-else class="space-y-6" @submit.prevent="onSubmit">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <RelationSelect
          v-model="form.brandId"
          relation="brands"
          :label="t('fields.brand')"
          :error="fieldErrors.brandId"
          required
        />
        <RelationMultiSelect
          v-model="form.categoryIds"
          relation="categories"
          :label="t('fields.categories')"
          :error="fieldErrors.categoryIds"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <LocalizedStringInput
          v-model="form.name"
          :locales="runtimeConfig.public.supportedLocales"
          required-locale="sr"
          :label="t('fields.name')"
          :error="fieldErrors.name"
          required
        />
        <LocalizedStringInput
          v-model="form.description"
          :locales="runtimeConfig.public.supportedLocales"
          required-locale="sr"
          :label="t('fields.description')"
          :error="fieldErrors.description"
        />
      </div>

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
            :items="getEnumOptions('ProductStatus').map((v) => ({ label: t(`enums.status.${v}`), value: v }))"
            class="w-full"
          />
        </UFormField>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <AttributeDetailsInput
          v-model="form.details"
          :locales="runtimeConfig.public.supportedLocales"
          :label="t('fields.details')"
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
          <UButton color="neutral" variant="soft" :to="config.route">{{ t('common.cancel') }}</UButton>
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
