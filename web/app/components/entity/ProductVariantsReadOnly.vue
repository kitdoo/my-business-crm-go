<script setup>
// Read-only variants/SKUs/stock summary for the Products view drawer
// (EntityViewDrawer.vue, config.view.variantsSummary) — same data as
// ProductVariantsPanel.vue / ProductSkusPanel.vue show on the full edit
// page, just flattened and non-interactive: no expand/collapse, no
// create/edit/delete affordances — this drawer is "look, don't touch",
// editing always happens on the full /products/:id page instead.
import { localizedText } from '~/utils/localizedText.js'
import { fetchTotalStockBySkuIds } from '~/utils/inventoryStock.js'

const props = defineProps({
  productId: { type: String, required: true },
})

const { t, locale } = useI18n()
const variantsApi = useEntityApi('productVariants')
const skusApi = useEntityApi('productSkus')
const inventoryApi = useEntityApi('inventory')
const { handle } = useApiErrorHandler()

const loading = ref(true)
const variants = ref([]) // [{ ...variant, skus: [{ ...sku, stock }] }]

function attributesSummary(attributes) {
  return Object.values(attributes || {})
    .map((v) => localizedText(v, locale.value))
    .filter(Boolean)
    .join(', ')
}

async function load() {
  loading.value = true
  try {
    const variantsRes = await variantsApi.list({
      filter: { productIds: [props.productId] },
      pagination: { limit: 100 },
    })
    const variantItems = variantsRes.items || []
    const variantIds = variantItems.map((variant) => variant.id)

    const skusRes = variantIds.length
      ? await skusApi.list({ filter: { variantIds }, pagination: { limit: 1000 } })
      : { items: [] }
    const skusByVariantId = {}
    for (const sku of skusRes.items || []) {
      ;(skusByVariantId[sku.variantId] ||= []).push(sku)
    }

    const stockBySkuId = await fetchTotalStockBySkuIds(
      inventoryApi,
      (skusRes.items || []).map((sku) => sku.id),
    )

    variants.value = variantItems.map((variant) => ({
      ...variant,
      skus: (skusByVariantId[variant.id] || []).map((sku) => ({ ...sku, stock: stockBySkuId[sku.id] || 0 })),
    }))
  } catch (err) {
    handle(err)
  } finally {
    loading.value = false
  }
}

watch(() => props.productId, load, { immediate: true })
</script>

<template>
  <div class="space-y-2">
    <h3 class="text-sm font-medium text-neutral-500">{{ t('entities.productVariants.label') }}</h3>
    <div v-if="loading" class="py-2 text-sm text-neutral-500">{{ t('common.loading') }}</div>
    <p v-else-if="variants.length === 0" class="text-sm text-neutral-500">{{ t('common.empty') }}</p>
    <div v-else class="space-y-3">
      <div v-for="variant in variants" :key="variant.id" class="rounded-md border border-default p-3 space-y-2">
        <div class="flex items-center gap-2 text-sm">
          <ImageThumbnail :value="variant.imageIds" />
          <span>{{ attributesSummary(variant.attributes) || t('entities.productVariants.label') }}</span>
        </div>
        <ul v-if="variant.skus.length" class="divide-y divide-neutral-100 rounded border border-neutral-100">
          <li v-for="sku in variant.skus" :key="sku.id" class="flex items-center justify-between px-2 py-1.5 text-sm">
            <span class="font-mono">{{ sku.sku }}</span>
            <span class="text-neutral-500">{{ t('entities.inventory.label') }}: {{ sku.stock }}</span>
          </li>
        </ul>
        <p v-else class="text-sm text-neutral-500">{{ t('common.empty') }}</p>
      </div>
    </div>
  </div>
</template>
