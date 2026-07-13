import { listActiveProducts } from '~~/server/utils/catalogClient.js'

const LOCALES = ['sr', 'en', 'ru'] // sr is unprefixed (defaultLocale), see nuxt.config.js i18n.strategy

// Per-locale slugs, mirroring nuxt.config.js i18n.pages exactly — sr is the
// page's own (Serbian) path, en/ru override with their own-language slug.
const STATIC_PAGES = [
  { sr: '/' },
  { sr: '/katalog', en: '/catalog', ru: '/katalog' },
  { sr: '/postani-diler', en: '/become-a-dealer', ru: '/stat-dilerom' },
  { sr: '/projekti', en: '/projects', ru: '/proekty' },
  { sr: '/kontakt', en: '/contact', ru: '/kontakt' },
  { sr: '/o-nama', en: '/about', ru: '/o-nas' },
]
const CATALOG_PREFIX = { sr: '/katalog', en: '/catalog', ru: '/katalog' }

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
    productSkus = items.map((p) => p.sku)
  } catch {
    // Catalog unreachable (e.g. guest account not provisioned yet) — still
    // emit the static pages rather than failing the whole sitemap.
  }

  const staticUrls = STATIC_PAGES.flatMap((page) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, page[locale] ?? page.sr)}</loc></url>`),
  )
  const productUrls = productSkus.flatMap((sku) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, `${CATALOG_PREFIX[locale]}/${sku}`)}</loc></url>`),
  )

  setHeader(event, 'Content-Type', 'application/xml')
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${[...staticUrls, ...productUrls].join('\n')}\n</urlset>`
})
