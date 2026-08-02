// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: true },

  app: {
    head: {
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/logos/logo_blue.png' },
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
    // A function, not a string: @nuxtjs/i18n calls this per-request
    // (see node_modules/@nuxtjs/i18n/dist/runtime/utils.js) rather than
    // resolving it once at `npm run build` time, so process.env here sees
    // the container's real env instead of the empty one Docker's build
    // stage has (same build-time-vs-runtime trap as gtag.id above).
    // Reading process.env.NUXT_PUBLIC_SITE_URL directly as a plain string
    // baked in an empty baseUrl, so useLocaleHead()'s canonical/hreflang
    // tags (app.vue) came out relative/wrong — this is why Search Console
    // was reporting pages as "Duplicate without user-selected canonical".
    // (A Nitro plugin copying runtimeConfig.public.siteUrl into
    // .i18n.baseUrl was tried instead and crashed the server — Nuxt
    // deep-freezes public runtimeConfig in production builds.)
    baseUrl: () => process.env.NUXT_PUBLIC_SITE_URL || '',
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
