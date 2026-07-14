<script setup>
// Movements tab on the ProductVariant detail page: the read-only ledger
// filtered to this variant, plus a "New movement" button that pre-fills
// variantId (EntityForm's initialValues) so the operator doesn't have to
// re-pick the variant they're already looking at.
const props = defineProps({
  variantId: { type: String, required: true },
})

const { t } = useI18n()
const { can } = usePermission()
const config = getEntityConfig('inventoryMovements')

const canCreate = computed(() => can(config.permissions.create))
const tableRef = ref(null)
const drawerOpen = ref(false)

function onSaved() {
  drawerOpen.value = false
  tableRef.value?.reload()
}
</script>

<template>
  <div class="space-y-4">
    <ProductStockSummary :variant-id="variantId" />

    <div class="flex justify-end">
      <UButton v-if="canCreate" icon="i-lucide-plus" @click="drawerOpen = true">
        {{ t('entities.inventoryMovements.create') }}
      </UButton>
    </div>
    <EntityDataTable ref="tableRef" entity="inventoryMovements" :fixed-filter="{ variantIds: [variantId] }" />

    <USlideover v-model:open="drawerOpen" side="right">
      <template #content>
        <div class="p-6 space-y-4 w-full max-w-md">
          <h2 class="text-lg font-semibold">{{ t('entities.inventoryMovements.create') }}</h2>
          <EntityForm
            entity="inventoryMovements"
            mode="drawer"
            :initial-values="{ variantId }"
            @saved="onSaved"
            @cancel="drawerOpen = false"
          />
        </div>
      </template>
    </USlideover>
  </div>
</template>
