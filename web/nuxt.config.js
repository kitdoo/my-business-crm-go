// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  // Opts into the app/ srcDir layout (pages, app.vue, app.config.js under
  // app/) — without this Nuxt 3 looks for them at the project root instead.
  future: { compatibilityVersion: 4 },

  // Node.js server runtime is required — @grpc/grpc-js needs net/http2,
  // not available on edge/serverless presets (see docs TD §1.1).
  nitro: {
    preset: 'node-server',
  },

  modules: [
    '@nuxt/ui',
    '@pinia/nuxt',
    '@nuxtjs/i18n',
  ],

  css: ['~/assets/css/main.css'],

  // Flat component names regardless of subdirectory (components/layout/
  // TopBar.vue -> <TopBar>, not <LayoutTopBar>) — matches how every
  // component is referenced across pages/components per the TD.
  components: [{ path: '~/components', pathPrefix: false }],

  // app/composables and app/utils are auto-imported by Nuxt by default;
  // app/config is not, but getEntityConfig/listEntityConfigs (EntityRegistry,
  // TD §9.1) are used everywhere without an explicit import — export them
  // here too.
  imports: {
    dirs: ['config', 'config/entities'],
  },

  i18n: {
    locales: [
      { code: 'sr', language: 'sr-RS', file: 'sr.json' },
      { code: 'en', language: 'en-US', file: 'en.json' },
      { code: 'ru', language: 'ru-RU', file: 'ru.json' },
    ],
    defaultLocale: 'sr',
    langDir: 'locales',
    strategy: 'no_prefix',
  },

  runtimeConfig: {
    // Private — server routes only, never reaches the browser bundle.
    grpc: {
      baseUrl: '',
      timeoutMs: 15000,
    },
    images: {
      baseUrl: '',
      maxSizeBytes: 10485760,
    },
    session: {
      secret: '',
      ttlSeconds: 60 * 60 * 12,
    },
    public: {
      appName: 'MyBusiness CRM',
      defaultLocale: 'sr',
      supportedLocales: ['sr', 'en', 'ru'],
      imagesMaxSizeBytes: 10485760,
    },
  },
})
