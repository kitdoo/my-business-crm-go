// @nuxtjs/i18n resolves canonical/hreflang URLs from runtimeConfig.public
// .i18n.baseUrl at request time (see node_modules/@nuxtjs/i18n/dist/runtime
// /utils.js), but the module bakes its own default for that field in at
// build time from nuxt.config.js's `i18n.baseUrl` — see the comment there.
// Overwrite it here, once at server startup, from runtimeConfig.public
// .siteUrl instead, which Nuxt's own runtimeConfig mechanism already
// resolves correctly from NUXT_PUBLIC_SITE_URL at container start.
export default defineNitroPlugin(() => {
  const config = useRuntimeConfig()
  config.public.i18n.baseUrl = config.public.siteUrl
})
