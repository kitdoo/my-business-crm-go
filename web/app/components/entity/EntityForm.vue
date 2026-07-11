<script setup>
// Generic create/edit form driven entirely by entity.form.fields (TD
// §9.1/§9.3): validation, FieldMask, and etag handling all come from
// useEntityForm — this component only renders inputs per field type.
import { getEnumOptions } from '~/config/enums.js'

const props = defineProps({
  entity: { type: String, required: true },
  id: { type: String, default: null },
  // 'page': navigates on save/delete (standalone /entity/id route).
  // 'drawer': emits events instead, so the host (EntityListPage's
  // left-side edit drawer, TD §8.3) decides what happens next.
  mode: { type: String, default: 'page' },
})
const emit = defineEmits(['saved', 'cancel', 'deleted'])

const { t } = useI18n()
const router = useRouter()
const config = getEntityConfig(props.entity)
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()

const { form, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm(props.entity, props.id)

const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))
const isDrawer = computed(() => props.mode === 'drawer')

onMounted(() => {
  if (!isCreate) load()
})

async function onSubmit() {
  try {
    const record = await save()
    if (isDrawer.value) {
      emit('saved', record)
      return
    }
    if (isCreate && record) {
      router.push(`${config.route}/${record.id}`)
    }
  } catch {
    // handled by useApiErrorHandler inside save()
  }
}

function onCancel() {
  if (isDrawer.value) emit('cancel')
}

async function onDelete() {
  const api = useEntityApi(props.entity)
  const { handle } = useApiErrorHandler()
  try {
    await api.remove(props.id, form.value.etag)
    if (isDrawer.value) {
      emit('deleted')
    } else {
      router.push(config.route)
    }
  } catch (err) {
    handle(err)
  }
}

function fieldsToRender() {
  return config.form.fields.filter((f) => !f.editOnly || !isCreate)
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <form v-else class="space-y-6" @submit.prevent="onSubmit">
      <FormGrid>
        <template v-for="field in fieldsToRender()" :key="field.key">
          <LocalizedStringInput
            v-if="field.type === 'localizedString'"
            v-model="form[field.key]"
            :locales="runtimeConfig.public.supportedLocales"
            :required-locale="'sr'"
            :label="t(field.label)"
            :error="fieldErrors[field.key]"
          />
          <UFormField v-else-if="field.type === 'enum'" :label="t(field.label)" :error="fieldErrors[field.key]">
            <USelect
              v-model="form[field.key]"
              :items="getEnumOptions(field.enum).map((v) => ({ label: t(`enums.status.${v}`), value: v }))"
              class="w-full"
            />
          </UFormField>
          <UFormField
            v-else-if="field.type === 'text'"
            :label="t(field.label)"
            :required="field.required"
            :error="fieldErrors[field.key]"
          >
            <UInput
              v-model="form[field.key]"
              class="w-full"
              :type="field.inputType || 'text'"
              :required="field.required"
              :maxlength="field.maxLength"
            />
          </UFormField>
          <UFormField
            v-else-if="field.type === 'number'"
            :label="t(field.label)"
            :required="field.required"
            :error="fieldErrors[field.key]"
          >
            <UInputNumber v-model="form[field.key]" class="w-full" :min="field.min" :max="field.max" />
          </UFormField>
          <RelationSelect
            v-else-if="field.type === 'relation'"
            v-model="form[field.key]"
            :relation="field.relation"
            :label="t(field.label)"
            :error="fieldErrors[field.key]"
            :required="field.required"
            :tree="field.tree"
            :exclude-id="field.tree && !isCreate ? props.id : null"
          />
        </template>
      </FormGrid>

      <div class="flex items-center justify-between">
        <UButton v-if="canDelete" color="error" variant="soft" @click="confirmDeleteOpen = true">
          {{ t('common.delete') }}
        </UButton>
        <div class="ml-auto flex gap-2">
          <UButton v-if="isDrawer" color="neutral" variant="soft" @click="onCancel">{{ t('common.cancel') }}</UButton>
          <UButton v-else color="neutral" variant="soft" :to="config.route">{{ t('common.cancel') }}</UButton>
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
