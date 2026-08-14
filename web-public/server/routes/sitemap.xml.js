import { listActiveProducts, listActiveVariantsForProducts, listActiveSkusForVariants, listActiveCategories } from '~~/server/utils/catalogClient.js'
import { slugify } from '~~/server/utils/slugify.js'

const LOCALES = ['sr', 'en', 'ru'] // sr is unprefixed (defaultLocale), see nuxt.config.js i18n.strategy

// Per-locale slugs, mirroring nuxt.config.js i18n.pages exactly — sr is the
// page's own (Serbian) path, en/ru override with their own-language slug.
const STATIC_PAGES = [
  { sr: '/' },
  { sr: '/katalog', en: '/catalog', ru: '/katalog' },
  { sr: '/postani-diler', en: '/become-a-dealer', ru: '/stat-dilerom' },
  { sr: '/projekti', en: '/projects', ru: '/proekty' },
  { sr: '/kontakt', en: '/contact', ru: '/kontakt' },
  { sr: '/faq', en: '/faq', ru: '/faq' },
]
const CATALOG_PREFIX = { sr: '/katalog', en: '/catalog', ru: '/katalog' }
// Mirrors nuxt.config.js i18n.pages['katalog-kategorija-category'].
const CATEGORY_PREFIX = { sr: '/katalog/kategorija', en: '/catalog/category', ru: '/katalog/kategorija' }

function localizedPath(locale, path) {
  return locale === 'sr' ? path : `/${locale}${path}`
}

// TZ §6.1 — sitemap includes every localized variant of every URL.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl || ''

  let productSkus = []
  try {
    const { items } = await listActiveProducts({ limit: 200 })

    // Product itself carries no sku (see catalogClient.js) — the catalog
    // page URL is keyed on the "representative" SKU: the first active SKU
    // of the first active variant, same rule as toCards() in
    // server/api/catalog/products.get.js. A product with no active variant
    // or SKU is simply left out of the sitemap.
    const variants = await listActiveVariantsForProducts(items.map((p) => p.id))
    const variantsByProduct = new Map()
    for (const v of variants) {
      const list = variantsByProduct.get(v.productId)
      if (list) list.push(v)
      else variantsByProduct.set(v.productId, [v])
    }

    const skus = await listActiveSkusForVariants(variants.map((v) => v.id))
    const skusByVariant = new Map()
    for (const s of skus) {
      const list = skusByVariant.get(s.variantId)
      if (list) list.push(s)
      else skusByVariant.set(s.variantId, [s])
    }

    productSkus = items
      .map((p) => {
        const representativeVariant = (variantsByProduct.get(p.id) || [])[0]
        const representativeSku = representativeVariant ? (skusByVariant.get(representativeVariant.id) || [])[0] : null
        return representativeSku?.sku
      })
      .filter(Boolean)
  } catch {
    // Catalog unreachable (e.g. guest account not provisioned yet) — still
    // emit the static pages rather than failing the whole sitemap.
  }

  let categorySlugs = []
  try {
    const categories = await listActiveCategories()
    // Same slug derivation as CatalogPage.vue's categorySlugOf — a name that
    // slugifies to '' (e.g. a script slugify() doesn't transliterate) falls
    // back to the category id instead of being dropped, so this stays in
    // sync with the actual URL the category page resolves to.
    categorySlugs = categories.map((c) => slugify(c.name?.values?.sr) || c.id)
  } catch {
    // Same degrade-gracefully rule as the product fetch above.
  }

  const staticUrls = STATIC_PAGES.flatMap((page) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, page[locale] ?? page.sr)}</loc></url>`),
  )
  const productUrls = productSkus.flatMap((sku) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, `${CATALOG_PREFIX[locale]}/${sku}`)}</loc></url>`),
  )
  const categoryUrls = categorySlugs.flatMap((slug) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, `${CATEGORY_PREFIX[locale]}/${slug}`)}</loc></url>`),
  )

  setHeader(event, 'Content-Type', 'application/xml')
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${[...staticUrls, ...categoryUrls, ...productUrls].join('\n')}\n</urlset>`
})
