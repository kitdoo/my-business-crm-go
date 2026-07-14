<script setup>
// Variants list on the Product page: every ProductVariant (sku,
// attributes, images, own price/stock) that belongs to this product.
// Editing an existing variant always opens its own full page
// (/product-variants/:id — General/Price/Movements tabs, see
// ProductVariantGeneralForm.vue). Creating one, though, stays right here:
// the inline form below the table saves in place and reappears empty so
// several variants can be added back-to-back without ever leaving the
// product.
const props = defineProps({
  productId: { type: String, required: true },
})

const { t } = useI18n()
const config = getEntityConfig('productVariants')
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))
const tableRef = ref(null)
const creating = ref(false)

function goToVariant(item) {
  navigateTo(`/product-variants/${item.id}`)
}

function onCreated() {
  creating.value = false
  tableRef.value?.reload()
}
</script>

<template>
  <div class="space-y-4">
    <EntityDataTable
      ref="tableRef"
      entity="productVariants"
      :fixed-filter="{ productIds: [productId] }"
      @view="goToVariant"
      @edit="goToVariant"
    >
      <template v-if="canCreate && !creating" #actions>
        <UButton icon="i-lucide-plus" color="neutral" variant="soft" @click="creating = true">
          {{ t('common.create') }}
        </UButton>
      </template>
    </EntityDataTable>

    <div v-if="creating" class="rounded-lg border border-default p-4">
      <ProductVariantGeneralForm
        :product-id="productId"
        embedded
        @created="onCreated"
        @cancel="creating = false"
      />
    </div>
  </div>
</template>
