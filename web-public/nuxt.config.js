// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  app: {
    head: {
      link: [
        // Square favicon (logo_blue.png alone is 342x1424 — non-square
        // images are ignored by Google's search-results favicon crawler).
        // favicon.ico stays for browsers/crawlers that request it by
        // convention regardless of these link tags; both derive from the
        // same icon-512.png source, browsers scale it down as needed.
        { rel: 'icon', href: '/favicon.ico', sizes: 'any' },
        { rel: 'icon', type: 'image/png', sizes: '512x512', href: '/icons/icon-512.png' },
        { rel: 'apple-touch-icon', href: '/icons/icon-512.png' },
      ],
      meta: [
        { name: 'google-site-verification', content: 'gYUYGB6P5gmYSi3WSCVRZUUw90BV1xoSeXF-kCdTMJ8' },
      ],
    },
  },

  future: { compatibilityVersion: 4 },

  // Node.js server runtime is required — @grpc/grpc-js needs net/http2,
  // not available on edge/serverless presets (see TZ_PHOMI_Public_Site §1.2,
  // same constraint as the admin web/ project).
  nitro: {
    preset: 'node-server',
    routeRules: {
      // Catalog + home barely change minute to minute; anonymous traffic can
      // spike higher than the admin panel ever sees (TZ §10).
      '/': { swr: 300 },
      '/katalog/**': { swr: 300 },
      '/projekti/**': { swr: 300 },
      // public/images/** has no build-time content hashing, so a file can in
      // principle be overwritten in place under the same name — avoid
      // `immutable`/a multi-year max-age for that reason. 30 days still
      // fixes the "missing cache lifetimes" PageSpeed finding for repeat
      // visits without risking long-lived stale images.
      '/images/**': { headers: { 'cache-control': 'public, max-age=2592000' } },
    },
  },

  modules: [
    '@nuxt/ui',
    '@nuxt/image',
    '@nuxtjs/i18n',
    'nuxt-gtag',
  ],

  // GA4, client-side only. `enabled` is a build-time module option (a
  // Nuxt module's setup() only runs once, while building) — reading
  // process.env.NUXT_PUBLIC_GTAG_ID here would bake in whatever value is
  // present at `npm run build` time, which is empty in the Docker image
  // (built once via CI/Dockerfile, before any deploy-time env is known —
  // see docker-compose.yml's web-public service), permanently disabling
  // the plugin no matter what NUXT_PUBLIC_GTAG_ID is set to at container
  // start. So `enabled` stays hardcoded true, and only `id` is left for
  // Nuxt's own runtimeConfig mechanism to fill in from that same env var
  // at container start (public.gtag.id <-> NUXT_PUBLIC_GTAG_ID) — an
  // empty id makes the plugin a no-op (see nuxt-gtag's resolveTags), so
  // local dev / previews without the var set still send nothing.
  gtag: {
    enabled: true,
  },

  css: ['~/assets/css/main.css'],

  // /api/images/:id is same-origin already-proxied (server/api/images/[id].get.js);
  // the default ipx provider treats it as a local filesystem path (no protocol)
  // and 404s (IPX_FILE_NOT_FOUND). Resizing is done by the backend itself
  // (?w= query param, see internal/transports/http/handlers/image), passed
  // through as plain URL query strings rather than ipx/ image modifiers —
  // so skip the provider entirely rather than configuring ipx.http.domains.
  image: {
    provider: 'none',
  },

  components: [{ path: '~/components', pathPrefix: false }],

  imports: {
    dirs: ['config'],
  },

  i18n: {
    locales: [
      { code: 'sr', language: 'sr-RS', file: 'sr.json' },
      { code: 'en', language: 'en-US', file: 'en.json' },
      { code: 'ru', language: 'ru-RU', file: 'ru.json' },
    ],
    defaultLocale: 'sr',
    langDir: 'locales',
    // sr (default) has no URL prefix, en/ru are prefixed — TZ §6.1.
    strategy: 'prefix_except_default',
    // Left at the module default. @nuxtjs/i18n's own canonical/hreflang
    // generation (useLocaleHead()) needs this, but it can only be a build
    // -time string or a function — and a function doesn't survive Nuxt's
    // public runtimeConfig (which gets serialized for the client and
    // frozen in production; Nuxt even warns "may not be able to be
    // serialized" at build time), so it silently came back empty at
    // runtime and canonical/hreflang links came out missing or relative
    // — this is why Search Console was reporting pages as "Duplicate
    // without user-selected canonical", including the homepage. Fixed by
    // not using useLocaleHead() at all: app.vue and every page build
    // their own canonical/hreflang/og tags via useSeoHead() from
    // runtimeConfig.public.siteUrl instead, which is a plain string and
    // resolves correctly from NUXT_PUBLIC_SITE_URL at runtime (see
    // useSeoHead.js).
    baseUrl: '',
    bundle: { optimizeTranslationDirective: false },
    // Module default reads per-page path overrides from definePageMeta —
    // 'config' makes it read the centralized `pages` map below instead.
    customRoutes: 'config',
    // Per-locale slugs: sr keeps the page's own (Serbian) path since it's
    // the primary local-SEO surface, en/ru get slugs in their own
    // language instead of inheriting the Serbian one verbatim. Keyed by
    // route name (kebab-case of the file path, "index" segments dropped).
    pages: {
      katalog: { en: '/catalog', ru: '/katalog' },
      'katalog-kategorija-category': { en: '/catalog/category/[category]', ru: '/katalog/kategorija/[category]' },
      'katalog-sku': { en: '/catalog/[sku]', ru: '/katalog/[sku]' },
      kontakt: { en: '/contact', ru: '/kontakt' },
      faq: { en: '/faq', ru: '/faq' },
      projekti: { en: '/projects', ru: '/proekty' },
      'postani-diler': { en: '/become-a-dealer', ru: '/stat-dilerom' },
      'postani-diler-hvala': { en: '/become-a-dealer/thank-you', ru: '/stat-dilerom/spasibo' },
    },
  },

  runtimeConfig: {
    // Private — server routes only, never reaches the browser bundle.
    grpc: {
      baseUrl: '',
      timeoutMs: 15000,
      // Static API key this frontend is registered under on the backend
      // (CRMConfig.NotificationClients) — required by
      // NotificationsService.Send (contact/dealer forms). env:
      // NUXT_GRPC_CLIENT_KEY.
      clientKey: '',
    },
    images: {
      baseUrl: '',
    },
    public: {
      appName: 'PHOMI Srbija',
      defaultLocale: 'sr',
      supportedLocales: ['sr', 'en', 'ru'],
      adminUrl: '',
      // Absolute origin, no trailing slash — used to build sitemap.xml URLs
      // and robots.txt's Sitemap: line. env: NUXT_PUBLIC_SITE_URL.
      siteUrl: '',
    },
  },
})
