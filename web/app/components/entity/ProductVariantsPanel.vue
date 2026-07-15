<script setup>
// Variants section of the Product page (see pages/products/[id].vue) — the
// whole create/edit hierarchy (Product -> Variant -> SKU -> Price/Stock)
// lives on this one page, no tabs, no navigation away. Each variant is a
// collapsible block: collapsed shows a read-only summary row, expanded
// shows its General form (locked/edit pattern, see
// ProductVariantGeneralForm.vue) plus its own SKUs section
// (ProductSkusPanel.vue) right underneath. Saving a new variant
// auto-expands it so the operator can start adding SKUs immediately,
// without hunting for it in the list first.
import { STATUS_COLOR_MAP } from '~/design/tokens.js'
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  productId: { type: String, required: true },
})

const { t, locale } = useI18n()
const config = getEntityConfig('productVariants')
const api = useEntityApi('productVariants')
const { handle } = useApiErrorHandler()
const { can } = usePermission()

const canCreate = computed(() => can(config.permissions.create))

const items = ref([])
const loading = ref(true)
const creating = ref(false)
const expandedId = ref(null)

function attributesSummary(item) {
  return Object.values(item.attributes || {})
    .map((v) => localizedText(v, locale.value))
    .filter(Boolean)
    .join(', ')
}

async function load() {
  loading.value = true
  try {
    const res = await api.list({
      filter: { productIds: [props.productId] },
      sort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_ASC' },
      pagination: { limit: 100 },
    })
    items.value = res.items || []
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

function toggle(item) {
  expandedId.value = expandedId.value === item.id ? null : item.id
}

function onCreated(record) {
  creating.value = false
  items.value = [...items.value, record]
  expandedId.value = record.id
}

function onDeleted(item) {
  items.value = items.value.filter((v) => v.id !== item.id)
  if (expandedId.value === item.id) expandedId.value = null
}

onMounted(load)
</script>

<template>
  <div class="space-y-4">
    <div v-if="loading" class="py-8 text-center text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <p v-if="items.length === 0 && !creating" class="text-sm text-neutral-500">{{ t('common.empty') }}</p>

      <div v-for="item in items" :key="item.id" class="rounded-lg border border-default">
        <button type="button" class="w-full flex items-center gap-3 p-3 text-left cursor-pointer" @click="toggle(item)">
          <ImageThumbnail :value="item.imageIds" />
          <span class="flex-1 text-sm">{{ attributesSummary(item) || t('entities.productVariants.label') }}</span>
          <StatusBadge :status="item.status" :map="STATUS_COLOR_MAP.productVariant" />
          <UIcon
            :name="expandedId === item.id ? 'i-lucide-chevron-up' : 'i-lucide-chevron-down'"
            class="size-4 text-neutral-400"
          />
        </button>

        <div v-if="expandedId === item.id" class="border-t border-default p-4 space-y-6">
          <ProductVariantGeneralForm :id="item.id" @deleted="onDeleted(item)" />

          <div class="border-t border-default pt-4">
            <h3 class="text-sm font-semibold mb-3">{{ t('entities.productSkus.label') }}</h3>
            <ProductSkusPanel :variant-id="item.id" />
          </div>
        </div>
      </div>
    </template>

    <div v-if="creating" class="rounded-lg border border-default p-4">
      <ProductVariantGeneralForm :product-id="productId" @created="onCreated" @cancel="creating = false" />
    </div>

    <UButton v-if="canCreate && !creating" icon="i-lucide-plus" color="neutral" variant="soft" @click="creating = true">
      {{ t('common.create') }}
    </UButton>
  </div>
</template>
