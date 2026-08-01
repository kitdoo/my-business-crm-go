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
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
</template>
