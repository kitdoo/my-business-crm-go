<script setup>
import { localizedText } from '~/utils/localizedText.js'

const props = defineProps({
  product: { type: Object, required: true },
})
const { t, locale } = useI18n()
const localePath = useLocalePath()
const { addItem, setQty, getQty, removeItem } = useCart()

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
const displayedVariants = computed(() => (props.product.variants || []).slice(0, MAX_VARIANT_SWATCHES))
const hiddenVariantsCount = computed(() => Math.max(0, (props.product.variants?.length || 0) - MAX_VARIANT_SWATCHES))
const selectedVariantIndex = ref(0)
const selectedSkuIndex = ref(0)
const activeVariant = computed(() => props.product.variants?.[selectedVariantIndex.value])
const activeSku = computed(() => activeVariant.value?.skus?.[selectedSkuIndex.value])

const activeImages = computed(() => {
  const ids = activeVariant.value?.imageIds
  return ids?.length ? ids.map((id) => `/api/images/${id}`) : ['/images/product-placeholder.svg']
})
function selectVariant(index) {
  selectedVariantIndex.value = index
  selectedSkuIndex.value = 0
  activeImageIndex.value = 0
}
function selectSku(index) {
  selectedSkuIndex.value = index
}
function skuLabel(sku) {
  return Object.values(sku.attributes || {}).map((v) => localizedText(v, locale.value)).filter(Boolean).join(' / ') || sku.sku
}

const activeImageIndex = ref(0)
function prevImage() {
  activeImageIndex.value = (activeImageIndex.value - 1 + activeImages.value.length) % activeImages.value.length
}
function nextImage() {
  activeImageIndex.value = (activeImageIndex.value + 1) % activeImages.value.length
}

const qtyInCart = computed(() => getQty(activeSku.value.sku))

// The catalog page renders cards before price/stock has loaded (see
// katalog/index.vue's staged fetch) — inStock is null on a SKU until the
// follow-up request resolves it. Adding to cart needs a real price, so it
// stays disabled until then instead of momentarily adding a null-priced item.
const stockKnown = computed(() => activeSku.value.inStock !== null)

function cartItem() {
  return { ...props.product, sku: activeSku.value.sku, price: activeSku.value.price }
}
function increment() {
  if (qtyInCart.value) setQty(activeSku.value.sku, qtyInCart.value + 1)
  else addItem(cartItem())
}
function decrement() {
  if (qtyInCart.value <= 1) removeItem(activeSku.value.sku)
  else setQty(activeSku.value.sku, qtyInCart.value - 1)
}
</script>

<template>
  <div class="rounded-lg border border-black/10 overflow-hidden group flex flex-col h-full">
    <NuxtLink :to="localePath(`/katalog/${activeSku.sku}`)" class="block relative aspect-square overflow-hidden bg-gray-50">
      <NuxtImg :src="activeImages[activeImageIndex]" alt="" loading="lazy" class="w-full h-full object-cover" />
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
        <h3 class="font-medium mb-1 line-clamp-2"><LocalizedText :value="product.name" /></h3>
      </NuxtLink>

      <div v-if="otherVariants.length" class="mt-2 flex gap-1.5 flex-wrap items-center">
        <button
          v-for="(variant, i) in displayedVariants"
          :key="i"
          type="button"
          class="w-7 h-7 rounded overflow-hidden border-2 shrink-0"
          :class="i === selectedVariantIndex ? 'border-brand-500' : 'border-transparent hover:border-black/20'"
          @click.stop.prevent="selectVariant(i)"
        >
          <img
            :src="variant.imageIds?.[0] ? `/api/images/${variant.imageIds[0]}` : '/images/product-placeholder.svg'"
            alt=""
            class="w-full h-full object-cover"
          />
        </button>
        <span v-if="hiddenVariantsCount" class="text-xs text-black/50 px-1">{{ t('catalog.moreVariants', { n: hiddenVariantsCount }) }}</span>
      </div>

      <div v-if="activeVariant.skus.length > 1" class="mt-2 flex gap-1.5 flex-wrap">
        <button
          v-for="(sku, i) in activeVariant.skus"
          :key="sku.sku"
          type="button"
          class="px-2 h-6 rounded border text-xs shrink-0"
          :class="i === selectedSkuIndex ? 'border-brand-500 text-brand-700' : 'border-black/15 text-black/60 hover:border-black/30'"
          @click.stop.prevent="selectSku(i)"
        >
          {{ skuLabel(sku) }}
        </button>
      </div>

      <div class="mt-auto pt-2">
        <div v-if="!stockKnown" class="h-5 w-16 rounded bg-black/5 animate-pulse" />
        <MoneyLabel v-else-if="activeSku.price" :amount="activeSku.price.amount" :currency="activeSku.price.currency" class="font-semibold text-brand-700" />

        <div class="mt-2 flex items-center justify-between gap-2">
          <div v-if="!stockKnown" class="h-5 w-20 rounded-full bg-black/5 animate-pulse" />
          <StatusBadge v-else :available="activeSku.inStock" />

          <div v-if="qtyInCart" class="flex items-center gap-2">
            <button class="w-6 h-6 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="decrement">
              <UIcon name="i-lucide-minus" class="w-3 h-3" />
            </button>
            <span class="text-sm font-medium w-4 text-center">{{ qtyInCart }}</span>
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
