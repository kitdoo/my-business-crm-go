<script setup>
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  product: { type: Object, required: true },
})
const { t, locale } = useI18n()
const localePath = useLocalePath()
const { setQty } = useCart()
const { sortIndicesByStock, firstInStockIndex } = useStockSort()

// Two-tier picker: first the visual Variant (color/texture — swatches),
// then, within it, the purchasable SKU (size/thickness — pills). Variants
// have no stable id of their own on the wire, so selection is by index;
// switching either tier swaps the preview photo/price/availability/cart
// target in place, without a navigation (each variant/SKU carries its own,
// see products.get.js toCard).
const otherVariants = computed(() => (props.product.variants || []).filter((_, i) => i !== selectedVariantIndex.value))
// Swatch row is capped at MAX_VARIANT_SWATCHES so a product with dozens of
// colors doesn't blow out the card's height — the rest collapse into a
// single "+N" chip instead of wrapping onto more rows.
const MAX_VARIANT_SWATCHES = 5
// In-stock variants/SKUs sort first so a shopper doesn't land on an
// out-of-stock option (or have it pushed out of the capped swatch row) while
// an in-stock one sits further down the unsorted list.
function variantHasStock(variant) { return (variant.skus || []).some((s) => s.inStock === true) }
const sortedVariantIndices = computed(() => sortIndicesByStock(props.product.variants || [], variantHasStock))
const displayedVariantIndices = computed(() => sortedVariantIndices.value.slice(0, MAX_VARIANT_SWATCHES))
const hiddenVariantsCount = computed(() => Math.max(0, (props.product.variants?.length || 0) - MAX_VARIANT_SWATCHES))

// The catalog grid paints cards before stock is known, then swaps in a new
// `product` object (see CatalogPage's `items`) once it resolves — the card
// instance survives that swap (keyed by product id), so a plain `ref` seeded
// once at setup would stay frozen on its pre-stock pick. Deriving the
// default as a computed instead means it always reflects the current
// `product` prop; `userPicked*` only overrides it once the shopper actually
// clicks something (cleared on variant switch so the new variant re-picks
// its own default, in-stock, SKU).
const defaultVariantIndex = computed(() => sortedVariantIndices.value[0] ?? 0)
const userPickedVariantIndex = ref(null)
const selectedVariantIndex = computed(() => userPickedVariantIndex.value ?? defaultVariantIndex.value)
const activeVariant = computed(() => props.product.variants?.[selectedVariantIndex.value])

// Land on an in-stock SKU by default — landing on skus[0] regardless of
// stock made a color swap look broken (badge says unavailable) even though
// other sizes/thicknesses within that same color were in stock.
const defaultSkuIndex = computed(() => firstInStockIndex(activeVariant.value?.skus || [], (s) => s.inStock))
const userPickedSkuIndex = ref(null)
const selectedSkuIndex = computed(() => userPickedSkuIndex.value ?? defaultSkuIndex.value)
const activeSku = computed(() => activeVariant.value?.skus?.[selectedSkuIndex.value])

// "generation" is a SKU-level characteristic (same color+size can exist at
// more than one generation — see products.get.js) — never a per-product
// value, so it must follow the active size selection, not stay fixed.
const activeGeneration = computed(() => activeSku.value?.attributes?.generation)

// w=800: at the grid's 4-column breakpoint cards render ~300-350px wide,
// so w=400 looked soft on retina (~2x) displays — see
// internal/transports/http/handlers/image for the fixed set of widths the
// backend will actually resize to.
const activeImages = computed(() => {
  const ids = activeVariant.value?.imageIds
  return ids?.length ? ids.map((id) => `/api/images/${id}?w=800`) : ['/images/product-placeholder.svg']
})
function selectVariant(index) {
  userPickedVariantIndex.value = index
  userPickedSkuIndex.value = null
  activeImageIndex.value = 0
}
function selectSku(index) {
  userPickedSkuIndex.value = index
}
const sortedSkuIndices = computed(() => sortIndicesByStock(activeVariant.value?.skus || [], (s) => s.inStock))
function skuLabel(sku) {
  // "generation" is surfaced as its own badge (see activeGeneration), not
  // repeated inline in the size pill.
  return Object.entries(sku.attributes || {})
    .filter(([key]) => key !== 'generation')
    .map(([, v]) => localizedText(v, locale.value))
    .filter(Boolean)
    .join(' / ') || sku.sku
}

const activeImageIndex = ref(0)
function prevImage() {
  activeImageIndex.value = (activeImageIndex.value - 1 + activeImages.value.length) % activeImages.value.length
}
function nextImage() {
  activeImageIndex.value = (activeImageIndex.value + 1) % activeImages.value.length
}

// The catalog page renders cards before price/stock has loaded (see
// katalog/index.vue's staged fetch) — inStock is null on a SKU until the
// follow-up request resolves it. Adding to cart needs a real price, so it
// stays disabled until then instead of momentarily adding a null-priced item.
const stockKnown = computed(() => activeSku.value.inStock !== null)

const productRef = computed(() => props.product)
const { qtyInCart, cartItem, addItem, increment, decrement } = useProductCartItem(productRef, activeVariant, activeSku)
</script>

<template>
  <div class="rounded-lg border border-black/10 overflow-hidden group flex flex-col h-full">
    <NuxtLink :to="localePath(`/katalog/${activeSku.sku}`)" class="block relative aspect-square overflow-hidden bg-gray-50">
      <NuxtImg :src="activeImages[activeImageIndex]" :alt="localizedText(product.name, locale)" loading="lazy" class="w-full h-full object-cover" />
      <template v-if="activeImages.length > 1">
        <button
          class="absolute left-1 top-1/2 -translate-y-1/2 w-7 h-7 flex items-center justify-center rounded-full bg-white/80 opacity-0 group-hover:opacity-100 transition-opacity"
          :aria-label="t('catalog.prevPage')"
          @click.stop.prevent="prevImage"
        >
          <UIcon name="i-lucide-chevron-left" class="w-4 h-4" />
        </button>
        <button
          class="absolute right-1 top-1/2 -translate-y-1/2 w-7 h-7 flex items-center justify-center rounded-full bg-white/80 opacity-0 group-hover:opacity-100 transition-opacity"
          :aria-label="t('catalog.nextPage')"
          @click.stop.prevent="nextImage"
        >
          <UIcon name="i-lucide-chevron-right" class="w-4 h-4" />
        </button>
        <div class="absolute bottom-1.5 left-1/2 -translate-x-1/2 flex gap-1">
          <span
            v-for="(url, i) in activeImages"
            :key="url"
            class="w-1.5 h-1.5 rounded-full"
            :class="i === activeImageIndex ? 'bg-white' : 'bg-white/50'"
          />
        </div>
      </template>
    </NuxtLink>

    <div class="p-4 flex flex-col flex-1">
      <NuxtLink :to="localePath(`/katalog/${activeSku.sku}`)">
        <h3 class="font-medium mb-1 line-clamp-2 flex items-center gap-1.5">
          <LocalizedText :value="product.name" />
          <span v-if="activeGeneration" class="text-xs font-normal text-black/40 shrink-0"><LocalizedText :value="activeGeneration" /></span>
        </h3>
      </NuxtLink>

      <div v-if="otherVariants.length" class="mt-2 flex gap-1.5 flex-wrap items-center">
        <button
          v-for="i in displayedVariantIndices"
          :key="i"
          type="button"
          class="w-7 h-7 rounded overflow-hidden border-2 shrink-0"
          :class="i === selectedVariantIndex ? 'border-brand-500' : 'border-transparent hover:border-black/20'"
          @click.stop.prevent="selectVariant(i)"
        >
          <img
            :src="product.variants[i].imageIds?.[0] ? `/api/images/${product.variants[i].imageIds[0]}?w=200` : '/images/product-placeholder.svg'"
            :alt="localizedText(product.name, locale)"
            class="w-full h-full object-cover"
          />
        </button>
        <span v-if="hiddenVariantsCount" class="text-xs text-black/50 px-1">{{ t('catalog.moreVariants', { n: hiddenVariantsCount }) }}</span>
      </div>

      <div v-if="activeVariant.skus.length > 1" class="mt-2 flex gap-1.5 flex-wrap">
        <!-- Stock isn't known yet (see stockKnown) — pills would sort/settle
             once it resolves, so a skeleton avoids showing (and clicking
             into) a size order that's about to reshuffle. -->
        <template v-if="!stockKnown">
          <span v-for="i in activeVariant.skus.length" :key="i" class="h-6 w-10 rounded bg-black/5 animate-pulse" />
        </template>
        <button
          v-for="i in sortedSkuIndices"
          v-else
          :key="activeVariant.skus[i].sku"
          type="button"
          class="px-2 h-6 rounded border text-xs shrink-0"
          :class="i === selectedSkuIndex ? 'border-brand-500 text-brand-700' : 'border-black/15 text-black/60 hover:border-black/30'"
          @click.stop.prevent="selectSku(i)"
        >
          {{ skuLabel(activeVariant.skus[i]) }}
        </button>
      </div>

      <div class="mt-auto pt-2">
        <div v-if="!stockKnown" class="h-5 w-16 rounded bg-black/5 animate-pulse" />
        <MoneyLabel
          v-else-if="activeSku.price"
          :amount="activeSku.price.amount"
          :currency="activeSku.price.currency"
          :unit="product.priceUnit"
          class="font-semibold text-brand-700"
        />

        <div class="mt-2 flex items-center justify-between gap-2">
          <div v-if="!stockKnown" class="h-5 w-20 rounded-full bg-black/5 animate-pulse" />
          <StatusBadge v-else :available="activeSku.inStock" />

          <div v-if="qtyInCart" class="flex items-center gap-2">
            <button class="w-6 h-6 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="decrement">
              <UIcon name="i-lucide-minus" class="w-3 h-3" />
            </button>
            <UInput
              :model-value="qtyInCart"
              type="number"
              min="1"
              class="w-12"
              :ui="{ base: 'text-center px-1' }"
              @update:model-value="(v) => setQty(activeSku.sku, Number(v))"
            />
            <button class="w-6 h-6 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="increment">
              <UIcon name="i-lucide-plus" class="w-3 h-3" />
            </button>
          </div>
          <button
            v-else
            class="text-xs uppercase tracking-wide text-brand-700 hover:text-brand-500 flex items-center gap-1 disabled:opacity-30"
            :disabled="!stockKnown"
            :aria-label="t('cart.addToCart')"
            @click.stop.prevent="addItem(cartItem())"
          >
            <UIcon name="i-lucide-shopping-cart" class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
