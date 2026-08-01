<script setup>
// Generic create/edit form driven entirely by entity.form.fields (TD
// §9.1/§9.3): validation, FieldMask, and etag handling all come from
// useEntityForm — per-field rendering lives in EntityFormField.vue.
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
  // Optional hook: async (record) => void, awaited right after a
  // successful create (mode="page" only) and before navigating to the
  // new record's page. For the one entity that needs a second backend
  // call as part of "creating" it (Product's required price — TD §12.6)
  // instead of baking that into this generic component. The callback
  // owns its own error display; if it throws, navigation is skipped so
  // the operator isn't swept away from a partially-finished creation.
  beforeCreateRedirect: { type: Function, default: null },
  // Optional extra validity gate from the host, ANDed with the form's
  // own field validation — for the #extra-fields slot content (e.g.
  // Product's required price isn't a form field, so nothing here
  // validates it on its own).
  extraValid: { type: Boolean, default: true },
})
const emit = defineEmits(['saved', 'cancel', 'deleted'])

const { t } = useI18n()
const router = useRouter()
const toast = useToast()
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
// config.form.deleteGuard is an optional (record) => boolean — when
// present and it returns false for the loaded record, Delete stays hidden
// regardless of permission (e.g. Users hides it for admin accounts, which
// the backend refuses to delete outright — see ErrUserAdminProtected).
const canDelete = computed(
  () =>
    !isCreate &&
    can(config.permissions.delete) &&
    (!config.form.deleteGuard || !record.value || config.form.deleteGuard(record.value)),
)
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
  if (!props.extraValid) return
  try {
    const record = await save()
    if (isDrawer.value) {
      emit('saved', record)
      return
    }
    if (isCreate && record) {
      if (props.beforeCreateRedirect) await props.beforeCreateRedirect(record)
      router.push(`${config.route}/${record.id}`)
      return
    }
    // Full-page edit stays on the same page after a successful save (no
    // navigation to react to), so it needs its own confirmation — without
    // this, clicking Save silently succeeds and looks like it did nothing.
    toast.add({ title: t('common.saved'), color: 'success' })
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

function visibleFields() {
  // editOnly: hidden while creating (e.g. Brand.status — not settable on
  // Create). createOnly: hidden while editing (e.g. User.password — set
  // once at Create, changed only through the separate ChangePassword flow).
  return config.form.fields.filter((f) => (!f.editOnly || !isCreate) && (!f.createOnly || isCreate))
}
// advanced (e.g. Partner/Client's invoicing-only legal fields) fields sit
// behind a closed-by-default "more information" disclosure so they don't
// add friction to the common create flow — see EntityFormField.vue.
const mainFields = computed(() => visibleFields().filter((f) => !f.advanced))
const advancedFields = computed(() => visibleFields().filter((f) => f.advanced))
const advancedOpen = ref(false)

// On-hand quantity for the currently picked skuId/warehouseId pair,
// reported by WarehouseStockSelect — only meaningful for InventoryMovement
// (its `quantity` field is the only one with capByStock set), but kept
// here rather than inside that field since it needs to reach a sibling
// field's UInputNumber min, not just its own.
const stockAvailable = ref(null)
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <form class="space-y-6" @submit.prevent="onSubmit">
        <FormGrid>
          <EntityFormField
            v-for="field in mainFields"
            :key="field.key"
            :field="field"
            :form="form"
            :error="fieldErrors[field.key]"
            :is-create="isCreate"
            :locales="runtimeConfig.public.supportedLocales"
            :stock-available="stockAvailable"
            @update:available="stockAvailable = $event"
          />
        </FormGrid>

        <UCollapsible v-if="advancedFields.length" v-model:open="advancedOpen">
          <UButton
            color="neutral"
            variant="ghost"
            size="sm"
            :icon="advancedOpen ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
            trailing
          >
            {{ t('common.moreInformation') }}
          </UButton>
          <template #content>
            <FormGrid class="pt-4">
              <EntityFormField
                v-for="field in advancedFields"
                :key="field.key"
                :field="field"
                :form="form"
                :error="fieldErrors[field.key]"
                :is-create="isCreate"
                :locales="runtimeConfig.public.supportedLocales"
                :stock-available="stockAvailable"
                @update:available="stockAvailable = $event"
              />
            </FormGrid>
          </template>
        </UCollapsible>

        <slot name="extra-fields" />

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
            <UButton type="submit" :disabled="!extraValid" :loading="saving">{{ t('common.save') }}</UButton>
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
