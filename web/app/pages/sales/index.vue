<script setup>
// Not <EntityListPage> — Sale has no generic form/drawer (TD §12.3), so
// "Create" is a link to the wizard page and rows navigate to the detail
// page instead of opening a drawer.
const config = getEntityConfig('sales')
const { t } = useI18n()
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">{{ t(config.label) }}</h1>
      <UButton v-if="canCreate" icon="i-lucide-plus" to="/sales/new">{{ t('entities.sales.create') }}</UButton>
    </div>
    <EntityDataTable entity="sales" :row-to="(item) => `/sales/${item.id}`" />
  </div>
</template>
