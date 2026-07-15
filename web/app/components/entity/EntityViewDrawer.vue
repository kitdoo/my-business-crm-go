<script setup>
// Read-only "just show me the record" view (TD §8.3) — a row click opens
// this before any edit UI. Renders config.list.columns as label/value
// pairs (same column -> component mapping EntityDataTable uses, from
// ~/utils/entityColumns.js) so a new column shows up here for free. The
// Edit button is the only way into editing: EntityListPage decides
// whether that means switching to the drawer's edit mode or navigating to
// a full page, based on the viewed entity's own config.detailPage.
// `relatedSales`, when the entity config declares one (Client/Partner),
// preloads that record's sales underneath — same `<EntityDataTable
// entity="sales" :row-to="...">` the main /sales page already uses, just
// scoped to this record via fixedFilter. `config.view.variantsSummary`
// (Product only) does the same for a read-only variants/SKUs/stock
// summary — see ProductVariantsReadOnly.vue.
import { columnComponent, columnProps } from '~/utils/entityColumns.js'

const props = defineProps({
  entity: { type: String, required: true },
  id: { type: String, required: true },
  relatedSales: { type: Function, default: null }, // (record) => fixedFilter
})
const emit = defineEmits(['edit', 'close'])

const { t } = useI18n()
const config = getEntityConfig(props.entity)
const api = useEntityApi(props.entity)
const { can } = usePermission()
const { handle } = useApiErrorHandler()

const loading = ref(true)
const record = ref(null)
const canUpdate = computed(() => can(config.permissions.update))

async function load() {
  loading.value = true
  record.value = null
  try {
    record.value = await api.get(props.id)
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

function plainValue(col, item) {
  return item[col.key]
}

watch(() => [props.entity, props.id], load, { immediate: true })
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-lg font-semibold">{{ t(config.label) }}</h2>

    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <dl v-else-if="record" class="space-y-3">
      <div v-for="col in config.list.columns" :key="col.key" class="flex items-start justify-between gap-4 text-sm">
        <dt class="text-neutral-500 shrink-0 select-text">{{ t(col.label) }}</dt>
        <dd class="text-right select-text">
          <component :is="columnComponent(col)" v-if="columnComponent(col)" v-bind="columnProps(col, record)" />
          <span v-else>{{ plainValue(col, record) }}</span>
        </dd>
      </div>
    </dl>

    <div v-if="relatedSales && record" class="space-y-2">
      <h3 class="text-sm font-medium text-neutral-500">{{ t('entities.sales.label') }}</h3>
      <EntityDataTable entity="sales" :fixed-filter="relatedSales(record)" :row-to="(item) => `/sales/${item.id}`" />
    </div>

    <ProductVariantsReadOnly v-if="config.view?.variantsSummary && record" :product-id="record.id" />

    <div class="flex justify-end gap-2">
      <UButton color="neutral" variant="soft" @click="emit('close')">{{ t('common.cancel') }}</UButton>
      <UButton v-if="canUpdate" @click="emit('edit')">{{ t('common.edit') }}</UButton>
    </div>
  </div>
</template>
