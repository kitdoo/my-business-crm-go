<script setup>
// Fully generic list page: table + "Create" button gated on
// entity.permissions.create (TD §9.3). Create and row-edit both open the
// same right-side drawer with <EntityForm mode="drawer"> (TD §8.3) instead
// of navigating to a separate page.
const props = defineProps({
  entity: { type: String, required: true },
})

const { t } = useI18n()
const config = getEntityConfig(props.entity)
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))

const tableRef = ref(null)
const drawerOpen = ref(false)
const editingId = ref(null)

function openCreate() {
  editingId.value = null
  drawerOpen.value = true
}

function openEdit(item) {
  editingId.value = item.id
  drawerOpen.value = true
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
    <EntityDataTable ref="tableRef" :entity="entity" @edit="openEdit" />

    <USlideover v-model:open="drawerOpen" side="right">
      <template #content>
        <div class="p-6 space-y-4 w-full max-w-md">
          <h2 class="text-lg font-semibold">
            {{ editingId ? t(`entities.${entity}.edit`) : t(`entities.${entity}.create`) }}
          </h2>
          <EntityForm
            :key="editingId ?? 'create'"
            :entity="entity"
            :id="editingId"
            mode="drawer"
            @saved="onSaved"
            @cancel="drawerOpen = false"
            @deleted="onDeleted"
          />
        </div>
      </template>
    </USlideover>
  </div>
</template>
