<script setup>
import { localizedText } from '~/utils/localizedText.js'

const { t, locale } = useI18n()
const localeHead = useLocaleHead()
const route = useRoute()
const router = useRouter()
const localePath = useLocalePath()
const { getProduct } = useCatalogApi()
const { addItem, setQty, getQty, removeItem } = useCart()

const { data: product, error } = await useAsyncData(`product-${route.params.sku}`, () => getProduct(route.params.sku))
if (error.value) {
  throw createError({ statusCode: 404, statusMessage: 'Not found' })
}

// Two-tier picker: first the visual Variant (color/texture — cubes), then,
// within it, the purchasable SKU (size/thickness — pills), auto-selected
// when there's only one. Both start on whichever pair matches the sku the
// page loaded with; switching either tier swaps price/availability/photo
// in place (each variant/SKU carries its own, see products/[sku].get.js)
// and updates the URL to the newly selected SKU without a refetch.
function findInitialIndices() {
  const variants = product.value?.variants || []
  for (let vi = 0; vi < variants.length; vi++) {
    const si = variants[vi].skus.findIndex((s) => s.sku === product.value.sku)
    if (si !== -1) return [vi, si]
  }
  return [0, 0]
}
const [initialVariantIndex, initialSkuIndex] = findInitialIndices()
const selectedVariantIndex = ref(initialVariantIndex)
const selectedSkuIndex = ref(initialSkuIndex)
const activeVariant = computed(() => product.value?.variants?.[selectedVariantIndex.value])
const activeSku = computed(() => activeVariant.value?.skus?.[selectedSkuIndex.value])
// "generation" is a SKU-level characteristic (same color+size can exist at
// more than one generation — see products/[sku].get.js) — never a
// per-product value, so it must follow the active size selection.
const activeGeneration = computed(() => activeSku.value?.attributes?.generation)

function selectVariant(index) {
  selectedVariantIndex.value = index
  selectedSkuIndex.value = 0
  activeImage.value = 0
}
function selectSku(index) {
  selectedSkuIndex.value = index
}
function skuLabel(sku) {
  // "generation" is surfaced as its own badge (see activeGeneration), not
  // repeated inline in the size pill.
  return Object.entries(sku.attributes || {})
    .filter(([key]) => key !== 'generation')
    .map(([, v]) => localizedText(v, locale.value))
    .filter(Boolean)
    .join(' / ') || sku.sku
}
watch(activeSku, (sku) => {
  if (sku && sku.sku !== route.params.sku) {
    router.replace(localePath(`/katalog/${sku.sku}`))
  }
})

const qtyInCart = computed(() => getQty(activeSku.value.sku))

// Combines variant attributes (e.g. color) and SKU attributes (e.g. size)
// into one label so the cart/checkout message spells out exactly which
// option the visitor picked, not just the SKU code.
function cartOptions() {
  const attrs = { ...(activeVariant.value?.attributes || {}), ...(activeSku.value?.attributes || {}) }
  return Object.values(attrs).map((v) => localizedText(v, locale.value)).filter(Boolean).join(' / ')
}
function cartItem() {
  return { ...product.value, sku: activeSku.value.sku, price: activeSku.value.price, options: cartOptions() }
}
function increment() {
  if (qtyInCart.value) setQty(activeSku.value.sku, qtyInCart.value + 1)
  else addItem(cartItem())
}
function decrement() {
  if (qtyInCart.value <= 1) removeItem(activeSku.value.sku)
  else setQty(activeSku.value.sku, qtyInCart.value - 1)
}

const activeImage = ref(0)
// Same image ids feed the big hero and the 64px thumbnail strip (see
// template) at different widths — see internal/transports/http/handlers/image
// for the fixed set of widths the backend will actually resize to.
const imageUrls = computed(() => (activeVariant.value?.imageIds?.length
  ? activeVariant.value.imageIds.map((id) => `/api/images/${id}?w=800`)
  : ['/images/product-placeholder.svg']))
const imageThumbUrls = computed(() => (activeVariant.value?.imageIds?.length
  ? activeVariant.value.imageIds.map((id) => `/api/images/${id}?w=200`)
  : ['/images/product-placeholder.svg']))
function prevImage() {
  activeImage.value = (activeImage.value - 1 + imageUrls.value.length) % imageUrls.value.length
}
function nextImage() {
  activeImage.value = (activeImage.value + 1) % imageUrls.value.length
}

const detailsEntries = computed(() => Object.entries(product.value?.details || {}))

function variantImageUrl(variant) {
  return variant.imageIds?.[0] ? `/api/images/${variant.imageIds[0]}?w=200` : '/images/product-placeholder.svg'
}

const productName = computed(() => product.value?.name?.values?.[locale.value] || product.value?.name?.values?.sr || '')
const productDescription = computed(() => (product.value?.description?.values?.[locale.value] || product.value?.description?.values?.sr || '').slice(0, 160))

useSeoMeta({
  title: () => `${productName.value} — ${t('seo.katalog.title')}`,
  description: () => productDescription.value,
  ogTitle: () => productName.value,
  ogDescription: () => productDescription.value,
  ogImage: () => useAbsoluteUrl(imageUrls.value[0]),
})

// Product structured data (schema.org) — price/currency come as basis
// points (see useFormat.js), hence the /100. Google's rich-result Product
// snippet and AI answer engines alike read this instead of the rendered
// price/stock text.
const productJsonLd = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'Product',
  name: productName.value,
  description: productDescription.value,
  image: imageUrls.value.map((url) => useAbsoluteUrl(url)),
  sku: activeSku.value?.sku,
  brand: { '@type': 'Brand', name: 'PHOMI' },
  offers: activeSku.value?.price ? {
    '@type': 'Offer',
    url: useAbsoluteUrl(route.fullPath),
    priceCurrency: activeSku.value.price.currency,
    price: (activeSku.value.price.amount / 100).toFixed(2),
    availability: activeSku.value.inStock ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock',
  } : undefined,
}))
useHead(() => ({
  link: localeHead.value.link,
  meta: localeHead.value.meta,
  script: [{ type: 'application/ld+json', innerHTML: JSON.stringify(productJsonLd.value) }],
}))

const contactHref = computed(() => `${localePath('/kontakt')}?message=${encodeURIComponent(t('catalog.contactPrefill', { sku: activeSku.value?.sku, name: productName.value }))}`)
</script>

<template>
  <div v-if="product" class="mx-auto max-w-5xl px-4 lg:px-8 py-12 lg:py-16">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-10">
      <div>
        <div class="relative aspect-square rounded-lg overflow-hidden bg-gray-50 mb-3 group">
          <NuxtImg :src="imageUrls[activeImage]" alt="" class="w-full h-full object-cover" />
          <template v-if="imageUrls.length > 1">
            <button
              class="absolute left-2 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full bg-white/80 opacity-0 group-hover:opacity-100 transition-opacity"
              :aria-label="t('catalog.prevPage')"
              @click="prevImage"
            >
              <UIcon name="i-lucide-chevron-left" class="w-5 h-5" />
            </button>
            <button
              class="absolute right-2 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full bg-white/80 opacity-0 group-hover:opacity-100 transition-opacity"
              :aria-label="t('catalog.nextPage')"
              @click="nextImage"
            >
              <UIcon name="i-lucide-chevron-right" class="w-5 h-5" />
            </button>
            <div class="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1.5">
              <span
                v-for="(url, i) in imageUrls"
                :key="url"
                class="w-2 h-2 rounded-full"
                :class="i === activeImage ? 'bg-white' : 'bg-white/50'"
              />
            </div>
          </template>
        </div>
        <div v-if="imageUrls.length > 1" class="flex gap-2 mb-3">
          <button
            v-for="(url, i) in imageThumbUrls"
            :key="url"
            class="w-16 h-16 rounded overflow-hidden border-2"
            :class="i === activeImage ? 'border-brand-500' : 'border-transparent'"
            @click="activeImage = i"
          >
            <img :src="url" alt="" class="w-full h-full object-cover" />
          </button>
        </div>

        <div v-if="product.variants.length > 1">
          <p class="text-sm text-black/50 mb-2">{{ t('catalog.otherVariants') }}</p>
          <div class="flex gap-3 flex-wrap">
            <button
              v-for="(variant, i) in product.variants"
              :key="i"
              type="button"
              class="w-20 h-20 rounded-lg overflow-hidden border-2 transition-colors"
              :class="i === selectedVariantIndex ? 'border-brand-500' : 'border-black/10 hover:border-brand-500'"
              @click="selectVariant(i)"
            >
              <img :src="variantImageUrl(variant)" alt="" class="w-full h-full object-cover" />
            </button>
          </div>
        </div>
      </div>

      <div>
        <p class="text-sm text-black/40 mb-1">{{ activeSku.sku }}</p>
        <h1 class="text-2xl font-bold mb-3 flex items-center gap-2">
          <LocalizedText :value="product.name" />
          <span v-if="activeGeneration" class="text-sm font-normal text-black/40"><LocalizedText :value="activeGeneration" /></span>
        </h1>
        <MoneyLabel
          v-if="activeSku.price"
          :amount="activeSku.price.amount"
          :currency="activeSku.price.currency"
          :unit="product.priceUnit"
          class="text-xl font-semibold text-brand-700 block mb-3"
        />
        <StatusBadge :available="activeSku.inStock" class="mb-6 inline-block" />

        <p class="text-black/70 leading-relaxed mb-6"><LocalizedText :value="product.description" /></p>

        <div v-if="activeVariant.skus.length > 1" class="mb-6">
          <p class="text-sm text-black/50 mb-2">{{ t('catalog.skuOptions') }}</p>
          <div class="flex gap-2 flex-wrap">
            <button
              v-for="(sku, i) in activeVariant.skus"
              :key="sku.sku"
              type="button"
              class="px-3 h-9 rounded-md border text-sm"
              :class="i === selectedSkuIndex ? 'border-brand-500 text-brand-700' : 'border-black/15 text-black/60 hover:border-black/30'"
              @click="selectSku(i)"
            >
              {{ skuLabel(sku) }}
            </button>
          </div>
        </div>

        <dl v-if="detailsEntries.length" class="space-y-2 mb-6 text-sm">
          <div v-for="[key, value] in detailsEntries" :key="key" class="flex justify-between border-b border-black/10 pb-2">
            <dt class="text-black/50">{{ key }}</dt>
            <dd><LocalizedText :value="value" /></dd>
          </div>
        </dl>

        <div class="flex flex-col sm:flex-row gap-3">
          <div v-if="qtyInCart" class="flex items-center justify-center gap-4 border border-brand-500 rounded-md py-4 flex-1">
            <button class="w-8 h-8 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="decrement">
              <UIcon name="i-lucide-minus" class="w-4 h-4" />
            </button>
            <UInput
              :model-value="qtyInCart"
              type="number"
              min="1"
              class="w-16"
              :ui="{ base: 'text-center px-1' }"
              @update:model-value="(v) => setQty(activeSku.sku, Number(v))"
            />
            <button class="w-8 h-8 flex items-center justify-center rounded-full border border-brand-500 text-brand-700" @click="increment">
              <UIcon name="i-lucide-plus" class="w-4 h-4" />
            </button>
          </div>
          <UButton v-else variant="cta-outline" size="xl" class="py-4 text-base flex-1" @click="addItem(cartItem())">
            {{ t('cart.addToCart') }}
          </UButton>
          <UButton :to="contactHref" variant="cta-outline" size="xl" class="py-4 text-base flex-1">
            {{ t('catalog.contactCta') }}
          </UButton>
        </div>
      </div>
    </div>
  </div>
</template>
