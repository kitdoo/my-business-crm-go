<script setup>
const { t, locale } = useI18n()
const localeHead = useLocaleHead()
const route = useRoute()
const localePath = useLocalePath()
const { getProduct } = useCatalogApi()

const { data: product, error } = await useAsyncData(`product-${route.params.sku}`, () => getProduct(route.params.sku))
if (error.value) {
  throw createError({ statusCode: 404, statusMessage: 'Not found' })
}

const activeImage = ref(0)
const imageUrls = computed(() => (product.value?.imageIds?.length
  ? product.value.imageIds.map((id) => `/api/images/${id}`)
  : ['/images/product-placeholder.svg']))

const detailsEntries = computed(() => Object.entries(product.value?.details || {}))

const productName = computed(() => product.value?.name?.values?.[locale.value] || product.value?.name?.values?.sr || '')
const productDescription = computed(() => (product.value?.description?.values?.[locale.value] || product.value?.description?.values?.sr || '').slice(0, 160))

useSeoMeta({
  title: () => `${productName.value} — ${t('seo.katalog.title')}`,
  description: () => productDescription.value,
  ogTitle: () => productName.value,
  ogDescription: () => productDescription.value,
  ogImage: () => imageUrls.value[0],
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))

const contactHref = computed(() => `${localePath('/kontakt')}?message=${encodeURIComponent(t('catalog.contactPrefill', { sku: product.value?.sku, name: productName.value }))}`)
</script>

<template>
  <div v-if="product" class="mx-auto max-w-5xl px-4 lg:px-8 py-12 lg:py-16">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-10">
      <div>
        <div class="aspect-square rounded-lg overflow-hidden bg-gray-50 mb-3">
          <NuxtImg :src="imageUrls[activeImage]" alt="" class="w-full h-full object-cover" />
        </div>
        <div v-if="imageUrls.length > 1" class="flex gap-2">
          <button
            v-for="(url, i) in imageUrls"
            :key="url"
            class="w-16 h-16 rounded overflow-hidden border-2"
            :class="i === activeImage ? 'border-brand-500' : 'border-transparent'"
            @click="activeImage = i"
          >
            <img :src="url" alt="" class="w-full h-full object-cover" />
          </button>
        </div>
      </div>

      <div>
        <p class="text-sm text-black/40 mb-1">{{ product.sku }}</p>
        <h1 class="text-2xl font-bold mb-3"><LocalizedText :value="product.name" /></h1>
        <MoneyLabel v-if="product.price" :amount="product.price.amount" :currency="product.price.currency" class="text-xl font-semibold text-brand-700 block mb-3" />
        <StatusBadge :status="product.status" class="mb-6 inline-block" />

        <p class="text-black/70 leading-relaxed mb-6"><LocalizedText :value="product.description" /></p>

        <dl v-if="detailsEntries.length" class="space-y-2 mb-8 text-sm">
          <div v-for="[key, value] in detailsEntries" :key="key" class="flex justify-between border-b border-black/10 pb-2">
            <dt class="text-black/50">{{ key }}</dt>
            <dd><LocalizedText :value="value" /></dd>
          </div>
        </dl>

        <UButton :to="contactHref" variant="cta-outline" size="xl" class="py-4 text-base" block>
          {{ t('catalog.contactCta') }}
        </UButton>
      </div>
    </div>
  </div>
</template>
