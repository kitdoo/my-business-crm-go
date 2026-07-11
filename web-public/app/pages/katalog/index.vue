<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()
const { listProducts, listCategories } = useCatalogApi()
const route = useRoute()
const router = useRouter()

const activeCategoryId = ref(route.query.category?.toString() || '')

const { data: categoriesData } = await useAsyncData('catalog-categories', () => listCategories())
const categories = computed(() => categoriesData.value?.items || [])

const items = ref([])
const nextCursor = ref(null)
const loading = ref(false)
const initialLoading = ref(true)

async function loadPage(reset = false) {
  loading.value = true
  try {
    const response = await listProducts({
      categoryId: activeCategoryId.value || undefined,
      cursor: reset ? undefined : nextCursor.value || undefined,
    })
    items.value = reset ? response.items : [...items.value, ...response.items]
    nextCursor.value = response.nextCursor
  } finally {
    loading.value = false
    initialLoading.value = false
  }
}

await loadPage(true)

watch(activeCategoryId, (val) => {
  router.replace({ query: { ...route.query, category: val || undefined } })
  loadPage(true)
})

useSeoMeta({
  title: t('seo.katalog.title'),
  description: t('seo.katalog.description'),
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-8">{{ t('nav.katalog') }}</h1>

    <div v-if="categories.length" class="flex flex-wrap gap-2 mb-8">
      <button
        class="px-4 py-1.5 rounded-full text-sm border"
        :class="!activeCategoryId ? 'bg-brand-500 text-white border-brand-500' : 'border-black/20 hover:border-brand-500'"
        @click="activeCategoryId = ''"
      >
        {{ t('catalog.allCategories') }}
      </button>
      <button
        v-for="c in categories"
        :key="c.id"
        class="px-4 py-1.5 rounded-full text-sm border"
        :class="activeCategoryId === c.id ? 'bg-brand-500 text-white border-brand-500' : 'border-black/20 hover:border-brand-500'"
        @click="activeCategoryId = c.id"
      >
        <LocalizedText :value="c.name" />
      </button>
    </div>

    <p v-if="!initialLoading && !items.length" class="text-black/50 py-16 text-center">
      {{ t('catalog.empty') }}
    </p>

    <div class="catalog-grid">
      <ProductCard v-for="p in items" :key="p.id" :product="p" />
    </div>

    <div v-if="nextCursor" class="text-center mt-10">
      <UButton variant="cta-outline" size="xl" class="px-10 py-4 text-base" :loading="loading" @click="loadPage(false)">
        {{ t('catalog.loadMore') }}
      </UButton>
    </div>
  </div>
</template>

<style scoped>
.catalog-grid {
  display: grid;
  gap: 24px;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}
@media (min-width: 640px) { .catalog-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (min-width: 1024px) { .catalog-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (min-width: 1280px) { .catalog-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
</style>
