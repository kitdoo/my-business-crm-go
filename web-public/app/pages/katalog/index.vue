<script setup>
import { localizedText } from '~/utils/localizedText.js'

const { t, locale } = useI18n()
const localeHead = useLocaleHead()
const { listProducts, listCategories } = useCatalogApi()
const route = useRoute()
const router = useRouter()

const PAGE_SIZE = 50

// Reka UI's SelectItem rejects an empty-string value outright (it's the
// sentinel Radix/Reka reserves for "no selection"), which was silently
// breaking the whole category <USelect> — and apparently left its portal
// stuck open, eating clicks on the nav links behind it. ALL_CATEGORIES is
// a non-empty sentinel, translated back to "no filter" wherever it's used.
const ALL_CATEGORIES = '__all__'

const activeCategoryId = ref(route.query.category?.toString() || ALL_CATEGORIES)
const activeSort = ref(route.query.sort?.toString() || 'in_stock')
// undefined = no stock filter, true = "Na stanju", false = "Po porudžbini".
const activeInStock = ref(route.query.inStock === 'true' ? true : route.query.inStock === 'false' ? false : undefined)

function toggleInStock(value) {
  activeInStock.value = activeInStock.value === value ? undefined : value
}

const { data: categoriesData } = await useAsyncData('catalog-categories', () => listCategories())
const categories = computed(() => categoriesData.value?.items || [])
const categoryOptions = computed(() => [
  { label: t('catalog.allCategories'), value: ALL_CATEGORIES },
  ...categories.value.map((c) => ({ label: localizedText(c.name, locale.value), value: c.id })),
])

const sortOptions = computed(() => [
  { label: t('catalog.sort.inStock'), value: 'in_stock' },
  { label: t('catalog.sort.newest'), value: 'newest' },
  { label: t('catalog.sort.nameAsc'), value: 'name_asc' },
  { label: t('catalog.sort.nameDesc'), value: 'name_desc' },
])

const items = ref([])
const total = ref(null)
const loading = ref(false)
const initialLoading = ref(true)
// Backend unreachable/erroring — shown instead of the empty-catalog state so
// visitors (and Googlebot) get a real page instead of the SSR crashing the
// whole route with an unstyled 500 (the sitemap route already degrades the
// same way for this same dependency, see server/routes/sitemap.xml.js).
const loadError = ref(false)
// Backend pagination is cursor-based (no offset), so arbitrary page jumps
// aren't possible — we keep the cursor of every page we've visited so
// Prev/Next can move one page at a time.
const cursorHistory = ref([undefined])
const currentPage = ref(1)
const nextCursor = ref(null)

// total is null while a stock filter is active (see products.get.js) — the
// numbered pager needs a real count, so it's only shown when total is known.
const totalPages = computed(() => (total.value ? Math.max(1, Math.ceil(total.value / PAGE_SIZE)) : 1))
// Only pages whose start cursor we've already resolved (by having visited
// the page before it) can be jumped to directly — see loadPage.
const knownPages = computed(() => Array.from({ length: cursorHistory.value.length }, (_, i) => i + 1))
const showPager = computed(() => (total.value != null ? totalPages.value > 1 : currentPage.value > 1 || !!nextCursor.value))

async function loadPage(page) {
  loading.value = true
  try {
    const response = await listProducts({
      categoryId: activeCategoryId.value === ALL_CATEGORIES ? undefined : activeCategoryId.value,
      sort: activeSort.value,
      inStock: activeInStock.value,
      cursor: cursorHistory.value[page - 1],
      limit: PAGE_SIZE,
    })
    items.value = response.items
    total.value = response.total ?? null
    nextCursor.value = response.nextCursor || null
    currentPage.value = page
    loadError.value = false
    if (nextCursor.value) cursorHistory.value[page] = nextCursor.value
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
    initialLoading.value = false
  }
}

function resetAndLoad() {
  cursorHistory.value = [undefined]
  loadPage(1)
}

await loadPage(1)

watch([activeCategoryId, activeSort, activeInStock], () => {
  router.replace({
    query: {
      ...route.query,
      category: activeCategoryId.value === ALL_CATEGORIES ? undefined : activeCategoryId.value,
      sort: activeSort.value !== 'in_stock' ? activeSort.value : undefined,
      inStock: activeInStock.value === undefined ? undefined : String(activeInStock.value),
    },
  })
  resetAndLoad()
})

useSeoMeta({
  title: t('seo.katalog.title'),
  description: t('seo.katalog.description'),
  ogTitle: t('seo.katalog.title'),
  ogDescription: t('seo.katalog.description'),
  ogImage: '/images/mini_house.jpg',
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-4">{{ t('nav.katalog') }}</h1>
    <p class="text-black/70 leading-relaxed max-w-3xl mb-8">{{ t('catalog.intro') }}</p>

    <div class="flex flex-wrap items-end gap-4 mb-8">
      <div class="w-full sm:w-64">
        <label class="block text-xs uppercase tracking-wide text-black/50 mb-1">{{ t('catalog.allCategories') }}</label>
        <USelect v-model="activeCategoryId" :items="categoryOptions" value-key="value" class="w-full" />
      </div>
      <div class="w-full sm:w-56">
        <label class="block text-xs uppercase tracking-wide text-black/50 mb-1">{{ t('catalog.sort.label') }}</label>
        <USelect v-model="activeSort" :items="sortOptions" value-key="value" class="w-full" />
      </div>
      <div class="flex gap-2">
        <button
          type="button"
          class="px-4 h-9 rounded-full text-sm font-medium uppercase tracking-wide transition-colors"
          :class="activeInStock === true ? 'bg-green-600 text-white' : 'bg-black/10 text-black/60 hover:bg-black/15'"
          @click="toggleInStock(true)"
        >
          {{ t('catalog.filter.inStock') }}
        </button>
        <button
          type="button"
          class="px-4 h-9 rounded-full text-sm font-medium uppercase tracking-wide transition-colors"
          :class="activeInStock === false ? 'bg-gray-500 text-white' : 'bg-black/10 text-black/60 hover:bg-black/15'"
          @click="toggleInStock(false)"
        >
          {{ t('catalog.filter.onOrder') }}
        </button>
      </div>
    </div>

    <p v-if="loadError" class="text-black/50 min-h-[40vh] flex items-center justify-center text-center">
      {{ t('catalog.loadError') }}
    </p>
    <p
      v-else-if="!initialLoading && !items.length"
      class="text-black/50 min-h-[40vh] flex items-center justify-center text-center"
    >
      {{ t('catalog.empty') }}
    </p>

    <div v-if="!loadError" class="catalog-grid">
      <ProductCard v-for="p in items" :key="p.id" :product="p" />
    </div>

    <nav v-if="!loadError && showPager" class="flex items-center justify-center gap-1.5 mt-10">
      <button
        class="w-9 h-9 flex items-center justify-center rounded-full border border-black/20 disabled:opacity-30 hover:border-brand-500"
        :disabled="currentPage <= 1 || loading"
        :aria-label="t('catalog.prevPage')"
        @click="loadPage(currentPage - 1)"
      >
        <UIcon name="i-lucide-chevron-left" class="w-4 h-4" />
      </button>

      <template v-if="total != null">
        <button
          v-for="p in knownPages"
          :key="p"
          class="w-9 h-9 rounded-full text-sm border"
          :class="p === currentPage ? 'bg-brand-500 text-white border-brand-500' : 'border-black/20 hover:border-brand-500'"
          :disabled="loading"
          @click="loadPage(p)"
        >
          {{ p }}
        </button>
        <span v-if="totalPages > knownPages.length" class="px-1 text-black/40">…</span>
      </template>
      <span v-else class="text-sm text-black/50">{{ currentPage }}</span>

      <button
        class="w-9 h-9 flex items-center justify-center rounded-full border border-black/20 disabled:opacity-30 hover:border-brand-500"
        :disabled="!nextCursor || loading"
        :aria-label="t('catalog.nextPage')"
        @click="loadPage(currentPage + 1)"
      >
        <UIcon name="i-lucide-chevron-right" class="w-4 h-4" />
      </button>
    </nav>
  </div>
</template>

<style scoped>
.catalog-grid {
  display: grid;
  gap: 24px;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}
@media (min-width: 640px) { .catalog-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (min-width: 1024px) { .catalog-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
</style>
