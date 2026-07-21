<script setup>
import { localizedText } from '~/utils/localizedText.js'

const { t, locale } = useI18n()
const localeHead = useLocaleHead()
const { listProducts, listCategories, getProductsStock } = useCatalogApi()
const route = useRoute()
const router = useRouter()

const PAGE_SIZE = 16

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

// lazy: true still blocks SSR on a direct hit (so visitors/Googlebot get
// full HTML), but stops client-side <NuxtLink> navigation from waiting on
// this fetch — the route change happens immediately and the page shows
// its own pending state (below) instead of freezing on the previous page.
const { data: categoriesData } = useAsyncData('catalog-categories', () => listCategories(), { lazy: true })
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

// Backend pagination is cursor-based (no offset), so arbitrary page jumps
// aren't possible — we keep the cursor of every page we've visited so
// Prev/Next can move one page at a time.
const cursorHistory = ref([undefined])
const currentPage = ref(1)

function fetchProductsPage() {
  return listProducts({
    categoryId: activeCategoryId.value === ALL_CATEGORIES ? undefined : activeCategoryId.value,
    sort: activeSort.value,
    inStock: activeInStock.value,
    cursor: cursorHistory.value[currentPage.value - 1],
    limit: PAGE_SIZE,
  })
}

const { data: pageResult, pending, error, refresh } = useAsyncData('catalog-products', fetchProductsPage, { lazy: true, watch: false })

// Unfiltered pages come back from products.get.js with every sku's
// price/inStock left null (see its includePriceStock) so the grid can
// paint image/name/variant pickers immediately; this overlays price/stock
// per skuId once the follow-up request (below) resolves, without
// re-rendering the whole grid from scratch. Cards from a stock-filtered
// page already carry real values and never hit that follow-up request.
const stockBySkuId = ref(new Map())

const items = computed(() => {
  const raw = pageResult.value?.items || []
  if (!stockBySkuId.value.size) return raw
  return raw.map((p) => ({
    ...p,
    variants: (p.variants || []).map((v) => ({
      ...v,
      skus: (v.skus || []).map((s) => {
        const stock = stockBySkuId.value.get(s.skuId)
        return stock ? { ...s, price: stock.price, inStock: stock.inStock } : s
      }),
    })),
  }))
})
const total = computed(() => pageResult.value?.total ?? null)
const nextCursor = computed(() => pageResult.value?.nextCursor || null)
// Backend unreachable/erroring — shown instead of the empty-catalog state so
// visitors (and Googlebot) get a real page instead of the SSR crashing the
// whole route with an unstyled 500 (the sitemap route already degrades the
// same way for this same dependency, see server/routes/sitemap.xml.js).
const loadError = computed(() => !!error.value)
// True only before the very first page has ever resolved — later refreshes
// (page/filter changes) show the skeleton via `pending` on the grid itself
// instead of blanking the whole section out.
const initialLoading = computed(() => pending.value && !pageResult.value)

watch(pageResult, async (val) => {
  if (val?.nextCursor) cursorHistory.value[currentPage.value] = val.nextCursor
  if (!val) return

  const pendingSkuIds = (val.items || []).flatMap((p) =>
    (p.variants || []).flatMap((v) => v.skus.filter((s) => s.inStock === null).map((s) => s.skuId)),
  )
  if (!pendingSkuIds.length) return

  try {
    const stockResult = await getProductsStock(pendingSkuIds)
    const map = new Map()
    for (const item of stockResult.items) map.set(item.skuId, item)
    stockBySkuId.value = map
  } catch {
    // Price/stock enrichment failing shouldn't take the whole page down —
    // cards just keep showing their loading placeholder (see ProductCard).
  }
}, { immediate: true })

// total is null while a stock filter is active (see products.get.js) — the
// numbered pager needs a real count, so it's only shown when total is known.
const totalPages = computed(() => (total.value ? Math.max(1, Math.ceil(total.value / PAGE_SIZE)) : 1))
// Only pages whose start cursor we've already resolved (by having visited
// the page before it) can be jumped to directly — see loadPage.
const knownPages = computed(() => Array.from({ length: cursorHistory.value.length }, (_, i) => i + 1))
const showPager = computed(() => (total.value != null ? totalPages.value > 1 : currentPage.value > 1 || !!nextCursor.value))

function loadPage(page) {
  currentPage.value = page
  stockBySkuId.value = new Map()
  refresh()
}

function resetAndLoad() {
  cursorHistory.value = [undefined]
  currentPage.value = 1
  stockBySkuId.value = new Map()
  refresh()
}

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
  ogImage: '/images/social-share.jpg',
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
      v-else-if="!initialLoading && !pending && !items.length"
      class="text-black/50 min-h-[40vh] flex items-center justify-center text-center"
    >
      {{ t('catalog.empty') }}
    </p>

    <div v-if="!loadError && initialLoading" class="catalog-grid">
      <div v-for="i in PAGE_SIZE" :key="i" class="rounded-lg border border-black/10 overflow-hidden animate-pulse">
        <div class="aspect-square bg-black/5" />
        <div class="p-4 space-y-2">
          <div class="h-4 bg-black/5 rounded w-3/4" />
          <div class="h-4 bg-black/5 rounded w-1/3" />
        </div>
      </div>
    </div>
    <div v-else-if="!loadError" class="catalog-grid" :class="{ 'opacity-50 pointer-events-none transition-opacity': pending }">
      <ProductCard v-for="p in items" :key="p.id" :product="p" />
    </div>

    <nav v-if="!loadError && showPager" class="flex items-center justify-center gap-1.5 mt-10">
      <button
        class="w-9 h-9 flex items-center justify-center rounded-full border border-black/20 disabled:opacity-30 hover:border-brand-500"
        :disabled="currentPage <= 1 || pending"
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
          :disabled="pending"
          @click="loadPage(p)"
        >
          {{ p }}
        </button>
        <span v-if="totalPages > knownPages.length" class="px-1 text-black/40">…</span>
      </template>
      <span v-else class="text-sm text-black/50">{{ currentPage }}</span>

      <button
        class="w-9 h-9 flex items-center justify-center rounded-full border border-black/20 disabled:opacity-30 hover:border-brand-500"
        :disabled="!nextCursor || pending"
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
