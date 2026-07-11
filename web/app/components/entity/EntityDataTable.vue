<script setup>
// Generic server-paginated (cursor) table (TD §9.3). Below `md` it renders
// as a card list instead of a horizontally-scrolled table (TD §6) — adding
// a column here means editing one entity config file, not this component.
// Row click / the edit icon both emit 'edit' so the host (EntityListPage)
// can open the left-side edit drawer; delete is handled locally with a
// confirm dialog and reloads the table in place.
import LocalizedText from '~/components/display/LocalizedText.vue'
import StatusBadge from '~/components/display/StatusBadge.vue'
import DateLabel from '~/components/display/DateLabel.vue'
import RelationLabel from '~/components/display/RelationLabel.vue'
import EnumLabel from '~/components/display/EnumLabel.vue'
import { STATUS_COLOR_MAP } from '~/design/tokens.js'

const props = defineProps({
  entity: { type: String, required: true },
  fixedFilter: { type: Object, default: () => ({}) },
})
const emit = defineEmits(['edit'])

const { t } = useI18n()
const config = getEntityConfig(props.entity)
const api = useEntityApi(props.entity)
const { handle } = useApiErrorHandler()
const { can } = usePermission()

const items = ref([])
const loading = ref(true)
const nextCursor = ref(null)
const filter = ref({})
const sort = ref(config.list.defaultSort)

const deleting = ref(false)
const confirmDeleteOpen = ref(false)
const pendingDelete = ref(null)

const canUpdate = computed(() => can(config.permissions.update))
const canDelete = computed(() => can(config.permissions.delete))
const showActions = computed(() => canUpdate.value || canDelete.value)

const COMPONENTS = { LocalizedText, StatusBadge, DateLabel, RelationLabel, EnumLabel }

function columnComponent(col) {
  return COMPONENTS[col.component]
}
function columnProps(col, item) {
  if (col.component === 'StatusBadge') return { status: item[col.key], map: STATUS_COLOR_MAP[col.statusMap] }
  if (col.component === 'LocalizedText') return { value: item[col.key] }
  if (col.component === 'DateLabel') return { value: item[col.key] }
  if (col.component === 'RelationLabel') return { value: item[col.key], relation: col.relation }
  if (col.component === 'EnumLabel') return { value: item[col.key] }
  return {}
}
function plainValue(col, item) {
  return item[col.key]
}

async function load(reset = true) {
  loading.value = true
  try {
    const params = {
      sort: sort.value,
      pagination: { limit: 25, cursor: reset ? undefined : nextCursor.value },
      filter: { ...filter.value, ...props.fixedFilter },
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

function onEdit(item) {
  if (!canUpdate.value) return
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
            class="border-b border-neutral-100"
            :class="canUpdate ? 'hover:bg-neutral-50 cursor-pointer' : ''"
            @click="onEdit(item)"
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
                  @click="onEdit(item)"
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
          class="rounded-lg border border-neutral-200 p-3 space-y-1"
          :class="canUpdate ? 'cursor-pointer' : ''"
          @click="onEdit(item)"
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
              @click="onEdit(item)"
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
