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
  // Pre-filled field values for create mode only (e.g. a "New movement"
  // button on Product's Movements tab pre-filling productId). Ignored
  // once editing an existing record.
  initialValues: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['saved', 'cancel', 'deleted'])

const { t } = useI18n()
const router = useRouter()
const config = getEntityConfig(props.entity)
const runtimeConfig = useRuntimeConfig()
const { can } = usePermission()
const api = useEntityApi(props.entity)
// Composables that rely on Vue's inject (useI18n/useRoute underneath) must
// be called synchronously at setup top-level — calling them lazily inside
// an async handler reached through a teleported <ConfirmDialog> loses the
// active component instance and throws "Must be called at the top of a
// setup function". Every handler below reuses this one instance.
const { handle } = useApiErrorHandler()

const { form, record, etag, loading, saving, fieldErrors, etagConflict, isCreate, load, save, reloadAfterConflict } =
  useEntityForm(props.entity, props.id, props.initialValues)

const confirmDeleteOpen = ref(false)
const canDelete = computed(() => !isCreate && can(config.permissions.delete))
const isDrawer = computed(() => props.mode === 'drawer')

// Declarative non-CRUD actions (TD §12.2, e.g. Warehouse.Deactivate) — an
// RPC that isn't part of the standard 5 and isn't a form field edit, so it
// can't go through useEntityForm.save(). Config-driven like everything
// else here: config.form.actions, not a bespoke Vue file per entity.
// visibleWhen reads `record` (the full loaded entity), not `form` — a
// gating field like Warehouse.status is deliberately absent from `form`.
const visibleActions = computed(() => {
  if (isCreate || !config.form.actions || !record.value) return []
  return config.form.actions.filter(
    (action) =>
      can(config.permissions[action.permission] ?? config.permissions.update) && action.visibleWhen(record.value),
  )
})
const pendingAction = ref(null)
const runningAction = ref(false)

async function onActionConfirm() {
  const action = pendingAction.value
  if (!action) return
  runningAction.value = true
  try {
    const updated = await $fetch(action.endpoint, {
      method: action.method || 'POST',
      body: { id: props.id, etag: etag.value },
    })
    if (isDrawer.value) {
      emit('saved', updated)
    } else {
      await load()
    }
  } catch (err) {
    handle(err)
  } finally {
    runningAction.value = false
    pendingAction.value = null
  }
}

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
  try {
    await api.remove(props.id, etag.value)
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
  // editOnly: hidden while creating (e.g. Brand.status — not settable on
  // Create). createOnly: hidden while editing (e.g. User.password — set
  // once at Create, changed only through the separate ChangePassword flow).
  return config.form.fields.filter((f) => (!f.editOnly || !isCreate) && (!f.createOnly || isCreate))
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <NuxtLink
        v-if="isDrawer && !isCreate && config.detailPage"
        :to="`${config.route}/${props.id}`"
        class="text-sm text-brand-700 hover:underline inline-flex items-center gap-1"
      >
        <UIcon name="i-lucide-external-link" class="size-4" />
        {{ t('common.openFullPage') }}
      </NuxtLink>

      <form class="space-y-6" @submit.prevent="onSubmit">
        <FormGrid>
          <template v-for="field in fieldsToRender()" :key="field.key">
            <LocalizedStringInput
              v-if="field.type === 'localizedString'"
              v-model="form[field.key]"
              :locales="runtimeConfig.public.supportedLocales"
              :required-locale="'sr'"
              :label="t(field.label)"
              :error="fieldErrors[field.key]"
              :required="field.required"
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
                :disabled="field.immutableOnEdit && !isCreate"
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
            <RelationMultiSelect
              v-else-if="field.type === 'relationMulti'"
              v-model="form[field.key]"
              :relation="field.relation"
              :label="t(field.label)"
              :error="fieldErrors[field.key]"
            />
            <KeyValueLocalizedInput
              v-else-if="field.type === 'keyValueLocalized'"
              v-model="form[field.key]"
              :locales="runtimeConfig.public.supportedLocales"
              :label="t(field.label)"
            />
            <UFormField v-else-if="field.type === 'images'" :label="t(field.label)" class="md:col-span-2">
              <ProductImageUploader v-model="form[field.key]" />
            </UFormField>
          </template>
        </FormGrid>

        <div class="flex items-center justify-between">
          <div class="flex gap-2">
            <UButton v-if="canDelete" color="error" variant="soft" @click="confirmDeleteOpen = true">
              {{ t('common.delete') }}
            </UButton>
            <UButton
              v-for="action in visibleActions"
              :key="action.key"
              color="warning"
              variant="soft"
              @click="pendingAction = action"
            >
              {{ t(action.label) }}
            </UButton>
          </div>
          <div class="ml-auto flex gap-2">
            <UButton v-if="isDrawer" color="neutral" variant="soft" @click="onCancel">{{ t('common.cancel') }}</UButton>
            <UButton v-else color="neutral" variant="soft" :to="config.route">{{ t('common.cancel') }}</UButton>
            <UButton type="submit" :loading="saving">{{ t('common.save') }}</UButton>
          </div>
        </div>
      </form>
    </template>

    <ConfirmDialog
      v-model:open="confirmDeleteOpen"
      :title="t('common.deleteConfirmTitle')"
      :description="t('common.deleteConfirmBody')"
      @confirm="onDelete"
    />
    <ConfirmDialog
      :open="!!pendingAction"
      :title="pendingAction ? t(pendingAction.confirmTitle) : ''"
      :description="pendingAction ? t(pendingAction.confirmBody) : ''"
      :danger="false"
      @update:open="(v) => !v && (pendingAction = null)"
      @confirm="onActionConfirm"
    />
    <EtagConflictDialog v-model:open="etagConflict" @reload="reloadAfterConflict" />
  </div>
</template>
