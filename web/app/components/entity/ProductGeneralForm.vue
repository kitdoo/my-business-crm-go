<script setup>
// Bespoke replacement for generic <EntityForm entity="products"> (TD
// §12.1/§12.6) — Product needs a grouped block layout (brand/categories,
// name/description, characteristics) that the generic field-by-field
// <FormGrid> loop can't express. sku/price/stock live on ProductSKU and
// images on ProductVariant (see ProductVariantGeneralForm.vue /
// ProductSkuGeneralForm.vue) — this form only ever handles the
// Product's own shared fields; variants are added afterwards inline on
// this same page (multiple, one at a time — see ProductVariantsPanel.vue).
//
// Existing products default to a locked (read-only) view with a pencil
// button to unlock editing — fields sit inside a native <fieldset
// disabled>, which reliably cascades to plain <input>/<button> elements,
// but custom dropdown components (RelationSelect, RelationMultiSelect,
// USelect) render their trigger as something the browser's native
// fieldset-disabling doesn't reach, so every one of those still needs its
// own explicit `:disabled="locked"` below. A brand-new product (no id
// yet) is never locked: there's nothing to look at until it's saved.
import { getEnumOptions } from '~/config/enums.js'

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
const { handle } = useApiErrorHandler()

const { form, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm('products', props.id, {})

const locked = ref(!isCreate)
const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))
const canUpdate = computed(() => !isCreate && can(config.permissions.update))

// Status is editOnly (see config/entities/products.js) so it never lands
// in `form` on create — Create has no status field server-side
// (ProductNew always starts Draft, see entities/product.go) — picking
// Active here means an immediate follow-up Update call in onSubmit.
const createStatus = ref('PRODUCT_STATUS_DRAFT')

onMounted(() => {
  if (!isCreate) load()
})

function onEdit() {
  locked.value = false
}

function onCancelEdit() {
  fieldErrors.value = {}
  load()
  locked.value = true
}

async function onSubmit() {
  try {
    const record = await save()
    if (!record) return

    if (isCreate) {
      let created = record
      if (createStatus.value !== 'PRODUCT_STATUS_DRAFT') {
        try {
          created = await api.update(record.id, { status: createStatus.value }, ['status'], record.etag)
        } catch (err) {
          handle(err)
          return
        }
      }
      router.push(`${config.route}/${created.id}`)
      return
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
      <div v-if="locked && canUpdate" class="flex items-center justify-end">
        <UButton icon="i-lucide-pencil" color="neutral" variant="soft" @click="onEdit">
          {{ t('common.edit') }}
        </UButton>
      </div>

      <fieldset :disabled="locked" class="space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <RelationSelect
            v-model="form.brandId"
            relation="brands"
            :label="t('fields.brand')"
            :error="fieldErrors.brandId"
            :disabled="locked"
            required
          />
          <RelationMultiSelect
            v-model="form.categoryIds"
            relation="categories"
            :label="t('fields.categories')"
            :error="fieldErrors.categoryIds"
            :disabled="locked"
          />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <LocalizedStringInput
            v-model="form.name"
            :locales="runtimeConfig.public.supportedLocales"
            required-locale="sr"
            :label="t('fields.name')"
            :error="fieldErrors.name"
            :disabled="locked"
            required
          />
          <LocalizedStringInput
            v-model="form.description"
            :locales="runtimeConfig.public.supportedLocales"
            required-locale="sr"
            :label="t('fields.description')"
            :error="fieldErrors.description"
            :disabled="locked"
          />
        </div>

        <UFormField v-if="!isCreate" :label="t('fields.status')" :error="fieldErrors.status" class="max-w-xs">
          <USelect
            v-model="form.status"
            :items="getEnumOptions('ProductStatus').map((v) => ({ label: t(`enums.status.${v}`), value: v }))"
            :disabled="locked"
            class="w-full"
          />
        </UFormField>
        <UFormField v-else :label="t('fields.status')" class="max-w-xs">
          <USelect
            v-model="createStatus"
            :items="[
              { label: t('enums.status.PRODUCT_STATUS_DRAFT'), value: 'PRODUCT_STATUS_DRAFT' },
              { label: t('enums.status.PRODUCT_STATUS_ACTIVE'), value: 'PRODUCT_STATUS_ACTIVE' },
            ]"
            class="w-full"
          />
        </UFormField>

        <AttributeDetailsInput
          v-model="form.details"
          :locales="runtimeConfig.public.supportedLocales"
          :label="t('fields.details')"
          :disabled="locked"
        />
      </fieldset>

      <div v-if="!locked" class="flex items-center justify-between">
        <UButton v-if="canDelete" color="error" variant="soft" @click="confirmDeleteOpen = true">
          {{ t('common.delete') }}
        </UButton>
        <div class="ml-auto flex gap-2">
          <UButton v-if="isCreate" color="neutral" variant="soft" :to="config.route">{{ t('common.cancel') }}</UButton>
          <UButton v-else color="neutral" variant="soft" @click="onCancelEdit">{{ t('common.cancel') }}</UButton>
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
