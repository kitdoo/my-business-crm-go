<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()
useSeoMeta({
  title: t('seo.about.title'),
  description: t('seo.about.description'),
  ogTitle: t('seo.about.title'),
  ogDescription: t('seo.about.description'),
  ogImage: '/images/showroom/1.png',
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-6">{{ t('about.title') }}</h1>
    <p class="text-black/70 leading-relaxed mb-4">{{ t('about.phomiText') }}</p>
    <p class="text-black/70 leading-relaxed">{{ t('about.tomStudioText') }}</p>
  </div>
</template>
