// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  app: {
    head: {
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/logos/logo_blue.png' },
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
    },
  },

  modules: [
    '@nuxt/ui',
    '@nuxt/image',
    '@nuxtjs/i18n',
  ],

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
    baseUrl: process.env.NUXT_PUBLIC_SITE_URL || '',
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
      'katalog-sku': { en: '/catalog/[sku]', ru: '/katalog/[sku]' },
      kontakt: { en: '/contact', ru: '/kontakt' },
      'o-nama': { en: '/about', ru: '/o-nas' },
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
