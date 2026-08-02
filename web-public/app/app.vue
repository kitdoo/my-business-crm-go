<script setup>
const { locale } = useI18n()
// @nuxtjs/i18n doesn't inject hreflang/canonical tags on its own — they
// have to be pulled from useLocaleHead() and merged in explicitly.
const i18nHead = useLocaleHead()
useHead({
  htmlAttrs: computed(() => ({
    lang: locale.value,
    ...i18nHead.value.htmlAttrs,
  })),
  link: computed(() => i18nHead.value.link),
  meta: computed(() => i18nHead.value.meta),
})

// Site-wide Organization/LocalBusiness structured data (schema.org, same
// contact details as kontakt/index.vue) — read by both classic rich
// results (sitelinks search box, knowledge panel) and AI answer engines
// (Google AI Overviews, Bing Copilot...), which prefer JSON-LD facts over
// parsing the rendered page. One block on every page, not per-locale,
// since the business itself doesn't change with the language.
useHead({
  script: [{
    type: 'application/ld+json',
    innerHTML: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'HomeGoodsStore',
      name: 'PHOMI Srbija',
      url: useAbsoluteUrl('/'),
      logo: useAbsoluteUrl('/images/logos/logo_blue.png'),
      image: useAbsoluteUrl('/images/logos/logo_blue.png'),
      telephone: '+381655632551',
      email: 'serbia@phomi.info',
      address: {
        '@type': 'PostalAddress',
        streetAddress: 'Veselina Masleše 56',
        addressLocality: 'Novi Sad',
        addressCountry: 'RS',
      },
      sameAs: ['https://www.instagram.com/phomi_serbia/'],
    }),
  }],
})
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
