// joinURL (ufo) collapses "https://phomi-rs.com" + "/" down to
// "https://phomi-rs.com" with no trailing slash — wrong for the homepage,
// which must canonicalize to "https://phomi-rs.com/" exactly. siteUrl
// never has a trailing slash (see nuxt.config.js) and every path from
// switchLocalePath already starts with "/", so plain concatenation is
// both correct and simpler than a URL-joining helper here.
function absoluteUrl(siteUrl, path) {
  return `${siteUrl}${path}`
}

// Replacement for @nuxtjs/i18n's useLocaleHead(): that composable derives
// canonical/hreflang/og:url from runtimeConfig.public.i18n.baseUrl, which
// has to be set to a function (see nuxt.config.js's i18n.baseUrl comment)
// to read NUXT_PUBLIC_SITE_URL at runtime instead of Docker build time —
// but Nuxt can't carry a function through its public runtimeConfig (it
// warns "may not be able to be serialized" at build and the value comes
// back empty/undefined in the built server), so baseUrl silently resolved
// to "" in production and every canonical/hreflang link came out missing
// or relative (e.g. href="/katalog" instead of an absolute URL) — which is
// why Search Console flagged pages as "Duplicate without user-selected
// canonical" even on the homepage. runtimeConfig.public.siteUrl doesn't
// have this problem: it's a plain string, resolved correctly at runtime by
// Nuxt's normal NUXT_PUBLIC_SITE_URL env override (same mechanism already
// relied on by useAbsoluteUrl, sitemap.xml, robots.txt). This composable
// builds the same link/meta shape useLocaleHead() did, from that value.
export function useSeoHead() {
  const { locale, locales } = useI18n()
  const route = useRoute()
  const switchLocalePath = useSwitchLocalePath()
  const { siteUrl, defaultLocale } = useRuntimeConfig().public

  return computed(() => {
    const link = []
    const meta = []

    // route.path (not switchLocalePath(locale.value), which carries the
    // current query string along) — katalog/index.vue's category/sort/
    // inStock filters are query params, e.g. /katalog?category=x&sort=y,
    // and each combination must canonicalize to the same clean /katalog to
    // avoid the exact "duplicate content, no canonical" signal this whole
    // fix addresses.
    const canonicalHref = absoluteUrl(siteUrl, route.path)
    link.push({ hid: 'canonical', rel: 'canonical', href: canonicalHref })
    meta.push({ hid: 'og-url', property: 'og:url', content: canonicalHref })

    let defaultHref
    for (const l of locales.value) {
      const language = typeof l === 'string' ? undefined : l.language
      const code = typeof l === 'string' ? l : l.code
      if (!language) continue
      const path = switchLocalePath(code)
      if (!path) continue
      const href = absoluteUrl(siteUrl, path)
      link.push({ hid: `alt-${language}`, rel: 'alternate', href, hreflang: language })
      if (code === defaultLocale) defaultHref = href
    }
    if (defaultHref) {
      link.unshift({ hid: 'alt-x-default', rel: 'alternate', href: defaultHref, hreflang: 'x-default' })
    }

    const currentLanguage = locales.value.find((l) => (typeof l === 'string' ? l : l.code) === locale.value)?.language
    if (currentLanguage) {
      meta.push({ hid: 'og-locale', property: 'og:locale', content: currentLanguage.replace(/-/g, '_') })
      for (const l of locales.value) {
        const language = typeof l === 'string' ? undefined : l.language
        if (language && language !== currentLanguage) {
          meta.push({ hid: `og-locale-alt-${language}`, property: 'og:locale:alternate', content: language.replace(/-/g, '_') })
        }
      }
    }

    return { link, meta }
  })
}
