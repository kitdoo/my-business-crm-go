// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

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
    '@pinia/nuxt',
    '@nuxtjs/i18n',
  ],

  css: ['~/assets/css/main.css'],

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
