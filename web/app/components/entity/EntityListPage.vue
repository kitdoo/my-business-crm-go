<script setup>
// Fully generic list page: table + "Create" button gated on
// entity.permissions.create (TD §9.3). A row click opens a right-side
// drawer in read-only "view" mode (<EntityViewDrawer>) with an Edit
// button inside it; only Create and that Edit button ever open the
// editable <EntityForm mode="drawer">. `viewOverride` lets a list show a
// *different* entity's view when a row is clicked — used by
// Inventory/InventoryMovements (TD §12.4) to open the Product view
// instead of their own read-only fields.
const props = defineProps({
  entity: { type: String, required: true },
  viewOverride: { type: Function, default: null }, // (item) => { entity, id }
  // Scopes the table to a parent record (e.g. a Product's variants) —
  // always merged into the list request's filter, regardless of the
  // entity's own declared UI filters. See EntityDataTable's fixedFilter.
  // Filter fields are usually plural (productIds: [id]), unlike the
  // singular create-time field (productId: id) — see createDefaults.
  fixedFilter: { type: Object, default: () => ({}) },
  // Seeds the create step with the parent's id so it isn't re-picked:
  // pre-fills the drawer form's initial values, or — when the entity uses
  // a full create page (config.detailPage) — is passed as that page's
  // query params instead, since there's no other way to hand it data.
  createDefaults: { type: Object, default: () => ({}) },
})

const { t } = useI18n()
const config = getEntityConfig(props.entity)
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))

const tableRef = ref(null)
const drawerOpen = ref(false)
const drawerMode = ref(null) // 'view' | 'edit' | 'create'
const activeEntity = ref(props.entity)
const activeId = ref(null)

function openCreate() {
  if (config.detailPage) {
    navigateTo({ path: `${config.route}/new`, query: props.createDefaults })
    return
  }
  activeEntity.value = props.entity
  activeId.value = null
  drawerMode.value = 'create'
  drawerOpen.value = true
}

function openView(item) {
  const target = props.viewOverride ? props.viewOverride(item) : { entity: props.entity, id: item.id }
  activeEntity.value = target.entity
  activeId.value = target.id
  drawerMode.value = 'view'
  drawerOpen.value = true
}

function openEditFor(entityKey, id) {
  const targetConfig = getEntityConfig(entityKey)
  if (targetConfig.detailPage) {
    drawerOpen.value = false
    navigateTo(`${targetConfig.route}/${id}`)
    return
  }
  activeEntity.value = entityKey
  activeId.value = id
  drawerMode.value = 'edit'
  drawerOpen.value = true
}

// Pencil-icon shortcut on the table always edits this list's own entity.
function onEditClick(item) {
  openEditFor(props.entity, item.id)
}

function onEditFromView() {
  openEditFor(activeEntity.value, activeId.value)
}

function onSaved() {
  drawerOpen.value = false
  tableRef.value?.reload()
}

function onDeleted() {
  drawerOpen.value = false
  tableRef.value?.reload()
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ t(config.label) }}</h1>
      <UButton v-if="canCreate" icon="i-lucide-plus" @click="openCreate">
        {{ t('common.create') }}
      </UButton>
    </div>
    <EntityDataTable ref="tableRef" :entity="entity" :fixed-filter="fixedFilter" @view="openView" @edit="onEditClick" />

    <USlideover v-model:open="drawerOpen" side="right">
      <template #content>
        <div
          class="p-6 space-y-4 w-full"
          :class="drawerMode === 'view' && getEntityConfig(activeEntity).view?.relatedSales ? 'max-w-2xl' : 'max-w-md'"
        >
          <EntityViewDrawer
            v-if="drawerMode === 'view'"
            :entity="activeEntity"
            :id="activeId"
            :related-sales="getEntityConfig(activeEntity).view?.relatedSales"
            @edit="onEditFromView"
            @close="drawerOpen = false"
          />
          <template v-else>
            <h2 class="text-lg font-semibold">
              {{ drawerMode === 'edit' ? t(`entities.${activeEntity}.edit`) : t(`entities.${activeEntity}.create`) }}
            </h2>
            <EntityForm
              :key="`${activeEntity}:${activeId ?? 'create'}`"
              :entity="activeEntity"
              :id="activeId"
              mode="drawer"
              :initial-values="drawerMode === 'create' ? createDefaults : {}"
              @saved="onSaved"
              @cancel="drawerOpen = false"
              @deleted="onDeleted"
            />
          </template>
        </div>
      </template>
    </USlideover>
  </div>
</template>
