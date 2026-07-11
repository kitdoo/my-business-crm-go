import { listActiveProducts } from '~~/server/utils/catalogClient.js'

const LOCALES = ['sr', 'en', 'ru'] // sr is unprefixed (defaultLocale), see nuxt.config.js i18n.strategy
const STATIC_PATHS = ['/', '/katalog', '/postani-diler', '/projekti', '/kontakt', '/o-nama']

function localizedPath(locale, path) {
  return locale === 'sr' ? path : `/${locale}${path}`
}

// TZ §6.1 — sitemap includes every localized variant of every URL.
export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl || ''

  let productPaths = []
  try {
    const { items } = await listActiveProducts({ limit: 200 })
    productPaths = items.map((p) => `/katalog/${p.sku}`)
  } catch {
    // Catalog unreachable (e.g. guest account not provisioned yet) — still
    // emit the static pages rather than failing the whole sitemap.
  }

  const allPaths = [...STATIC_PATHS, ...productPaths]

  const urls = allPaths.flatMap((path) =>
    LOCALES.map((locale) => `  <url><loc>${siteUrl}${localizedPath(locale, path)}</loc></url>`),
  )

  setHeader(event, 'Content-Type', 'application/xml')
  return `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.join('\n')}\n</urlset>`
})
