<script setup>
// Generic server-paginated (cursor) table (TD §9.3). Below `md` it renders
// as a card list instead of a horizontally-scrolled table (TD §6) — adding
// a column here means editing one entity config file, not this component.
// Row click emits 'view' so the host (EntityListPage) opens a read-only
// view drawer with an Edit button inside it; the pencil icon is a direct
// shortcut that emits 'edit' straight into the edit drawer/page. Column
// -> display-component mapping lives in ~/utils/entityColumns.js, shared
// with EntityViewDrawer so both render a column's value identically. For
// entities with no generic edit at all (Sale, TD §12.3 — items immutable,
// status moves only through UpdateStatus/Cancel), pass `rowTo` instead:
// row click navigates to a full page rather than emitting 'view'.
import { columnComponent, columnProps } from '~/utils/entityColumns.js'

const props = defineProps({
  entity: { type: String, required: true },
  fixedFilter: { type: Object, default: () => ({}) },
  rowTo: { type: Function, default: null }, // (item) => route path
})
const emit = defineEmits(['view', 'edit'])

const { t } = useI18n()
const config = getEntityConfig(props.entity)
const api = useEntityApi(props.entity)
const { handle } = useApiErrorHandler()
const { can } = usePermission()

const items = ref([])
const loading = ref(true)
const nextCursor = ref(null)
const filter = ref({ ...(config.list.defaultFilter || {}) })
const sort = ref(config.list.defaultSort)

const deleting = ref(false)
const confirmDeleteOpen = ref(false)
const pendingDelete = ref(null)

// canUpdate gates only the pencil-icon shortcut; when `rowTo` is set
// (Sale) there's no edit drawer at all regardless of the update
// permission, so the pencil never renders.
const canUpdate = computed(() => !props.rowTo && can(config.permissions.update))
const canDelete = computed(() => can(config.permissions.delete))
const showActions = computed(() => canUpdate.value || canDelete.value)

function plainValue(col, item) {
  return item[col.key]
}

// `select`-type filters (single value in the UI) back a repeated proto
// field (e.g. Product.categoryIds) — wrap to a one-element array on the
// way out so EntityFilterBar can stay a plain single-select.
// `numberRange` (one compact min–max control in the UI) backs two
// separate wire fields (e.g. Inventory's minQuantity/maxQuantity) — split
// {min,max} into filterDef.minKey/maxKey. Empty/null values are dropped
// either way so an unset filter never sends `{key: []}` or `{key: null}`.
function filterPayload() {
  const payload = { ...filter.value }
  for (const filterDef of config.list.filters || []) {
    const value = payload[filterDef.key]
    if (filterDef.type === 'select' && value != null) {
      payload[filterDef.key] = [value]
      continue
    }
    if (filterDef.type === 'numberRange') {
      delete payload[filterDef.key]
      if (value?.min != null) payload[filterDef.minKey] = value.min
      if (value?.max != null) payload[filterDef.maxKey] = value.max
      continue
    }
    if (value == null || (Array.isArray(value) && value.length === 0)) {
      delete payload[filterDef.key]
    }
  }
  return payload
}

async function load(reset = true) {
  loading.value = true
  try {
    const params = {
      sort: sort.value,
      pagination: { limit: 25, cursor: reset ? undefined : nextCursor.value },
      filter: { ...filterPayload(), ...props.fixedFilter },
    }
    const res = await api.list(params)
    items.value = reset ? res.items : [...items.value, ...res.items]
    nextCursor.value = res.nextCursor || null
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

watch(filter, () => load(true), { deep: true })
watch(sort, () => load(true), { deep: true })

const sortItems = computed(() => (config.list.sort || []).map((s) => ({ label: t(s.label), value: s.field })))

function onSortFieldChange(field) {
  sort.value = { field, direction: sort.value?.direction || 'SORT_DIRECTION_DESC' }
}

function toggleSortDirection() {
  sort.value = {
    ...sort.value,
    direction: sort.value?.direction === 'SORT_DIRECTION_ASC' ? 'SORT_DIRECTION_DESC' : 'SORT_DIRECTION_ASC',
  }
}

// Every row is clickable: `rowTo` navigates away, otherwise a row click
// always opens the read-only view drawer (viewing needs only `read`,
// which is why the row is on screen at all — no update permission gate).
function onRowClick(item) {
  if (props.rowTo) {
    navigateTo(props.rowTo(item))
    return
  }
  emit('view', item)
}

function onEditClick(item) {
  emit('edit', item)
}

function onDeleteClick(item) {
  pendingDelete.value = item
  confirmDeleteOpen.value = true
}

async function onDeleteConfirm() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await api.remove(pendingDelete.value.id, pendingDelete.value.etag)
    await load(true)
  } catch (err) {
    handle(err)
  } finally {
    deleting.value = false
    pendingDelete.value = null
  }
}

defineExpose({ reload: () => load(true) })
onMounted(() => load(true))
</script>

<template>
  <div>
    <div v-if="config.list.filters?.length || sortItems.length" class="flex flex-wrap items-end justify-between gap-3 mb-4">
      <EntityFilterBar v-if="config.list.filters?.length" :filters="config.list.filters" v-model="filter" />

      <div v-if="sortItems.length" class="flex items-end gap-1">
        <UFormField :label="t('sort.label')">
          <USelectMenu
            :model-value="sort?.field"
            :items="sortItems"
            value-key="value"
            class="w-40"
            @update:model-value="onSortFieldChange"
          />
        </UFormField>
        <UButton
          :icon="sort?.direction === 'SORT_DIRECTION_ASC' ? 'i-lucide-arrow-up-narrow-wide' : 'i-lucide-arrow-down-wide-narrow'"
          color="neutral"
          variant="soft"
          :aria-label="t('sort.direction')"
          @click="toggleSortDirection"
        />
      </div>
    </div>

    <div v-if="loading && items.length === 0" class="py-8 text-center text-neutral-500">
      {{ t('common.loading') }}
    </div>
    <div v-else-if="items.length === 0" class="py-8 text-center text-neutral-500">
      {{ t('common.empty') }}
    </div>
    <template v-else>
      <!-- Table: >= md -->
      <table class="hidden md:table w-full text-sm">
        <thead>
          <tr class="border-b border-neutral-200 text-left">
            <th v-for="col in config.list.columns" :key="col.key" class="py-2 px-3 font-medium text-neutral-600">
              {{ t(col.label) }}
            </th>
            <th v-if="showActions" class="py-2 px-3 font-medium text-neutral-600">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="item.id"
            class="border-b border-neutral-100 hover:bg-neutral-50 cursor-pointer"
            @click="onRowClick(item)"
          >
            <td v-for="col in config.list.columns" :key="col.key" class="py-2 px-3">
              <component :is="columnComponent(col)" v-if="columnComponent(col)" v-bind="columnProps(col, item)" />
              <span v-else>{{ plainValue(col, item) }}</span>
            </td>
            <td v-if="showActions" class="py-2 px-3" @click.stop>
              <div class="flex items-center gap-1">
                <UButton
                  v-if="canUpdate"
                  icon="i-lucide-pencil"
                  color="primary"
                  variant="ghost"
                  size="xs"
                  :aria-label="t('common.edit')"
                  @click="onEditClick(item)"
                />
                <UButton
                  v-if="canDelete"
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="xs"
                  :aria-label="t('common.delete')"
                  @click="onDeleteClick(item)"
                />
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Cards: < md -->
      <div class="md:hidden space-y-2">
        <div
          v-for="item in items"
          :key="item.id"
          class="rounded-lg border border-neutral-200 p-3 space-y-1 cursor-pointer"
          @click="onRowClick(item)"
        >
          <div v-for="col in config.list.columns" :key="col.key" class="flex justify-between text-sm">
            <span class="text-neutral-500">{{ t(col.label) }}</span>
            <component :is="columnComponent(col)" v-if="columnComponent(col)" v-bind="columnProps(col, item)" />
            <span v-else>{{ plainValue(col, item) }}</span>
          </div>
          <div v-if="showActions" class="flex items-center justify-end gap-1 pt-1" @click.stop>
            <UButton
              v-if="canUpdate"
              icon="i-lucide-pencil"
              color="primary"
              variant="ghost"
              size="xs"
              :aria-label="t('common.edit')"
              @click="onEditClick(item)"
            />
            <UButton
              v-if="canDelete"
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="xs"
              :aria-label="t('common.delete')"
              @click="onDeleteClick(item)"
            />
          </div>
        </div>
      </div>
    </template>

    <div v-if="nextCursor" class="pt-4 text-center">
      <UButton variant="soft" :loading="loading" @click="load(false)">{{ t('common.loadMore') }}</UButton>
    </div>

    <ConfirmDialog
      v-model:open="confirmDeleteOpen"
      :title="t('common.deleteConfirmTitle')"
      :description="t('common.deleteConfirmBody')"
      @confirm="onDeleteConfirm"
    />
  </div>
</template>
