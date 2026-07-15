<script setup>
// Bespoke replacement for generic <EntityForm entity="productVariants">
// (mirrors ProductGeneralForm.vue's rationale) — a variant groups
// attributes/images together and needs the locked/edit pattern below.
// Carries no sku/price of its own (see ProductSkuGeneralForm.vue for
// those). No Status field: a variant is always Active the moment it's
// saved (flipped right after create, same as ProductSkuGeneralForm) —
// the only other lifecycle action is Delete, never a manual draft/inactive
// toggle. Always rendered inline on the Product page (see
// ProductVariantsPanel.vue) — there is no standalone /product-variants/:id
// page. productId comes from the parent Product's page as a create-time
// prop (createOnly — see useEntityForm's applyRecord — so it's never part
// of the reactive `form`, only sent on Create) and is otherwise never
// touched again. This component never navigates: every outcome (created/cancelled/
// deleted) is emitted for ProductVariantsPanel to fold into its own list
// state in place.
const props = defineProps({
  id: { type: String, default: null },
  productId: { type: String, default: null },
})
const emit = defineEmits(['created', 'cancel', 'deleted'])

const { t } = useI18n()
const toast = useToast()
const config = getEntityConfig('productVariants')
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()
const api = useEntityApi('productVariants')
const { handle } = useApiErrorHandler()

const { form, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm('productVariants', props.id, { productId: props.productId })

// Same locked/edit pattern as ProductGeneralForm.vue — an existing variant
// defaults to read-only with a pencil button to unlock. Never applies to
// create: there's nothing to look at yet.
const locked = ref(!isCreate)
const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))
const canUpdate = computed(() => !isCreate && can(config.permissions.update))

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
      // Create has no status field server-side (ProductVariantNew always
      // starts Draft, see entities/product_variant.go) — a variant is
      // usable as soon as it's saved, so flip it to Active right away
      // instead of leaving it stuck in Draft.
      let created = record
      try {
        created = await api.update(record.id, { status: 'PRODUCT_VARIANT_STATUS_ACTIVE' }, ['status'], record.etag)
      } catch (err) {
        handle(err)
        return
      }
      emit('created', created)
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
    emit('deleted')
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
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <AttributeDetailsInput
            v-model="form.attributes"
            :locales="runtimeConfig.public.supportedLocales"
            :label="t('fields.attributes')"
            :disabled="locked"
          />
          <UFormField :label="t('fields.images')">
            <ProductImageUploader v-model="form.imageIds" />
          </UFormField>
        </div>
      </fieldset>

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
