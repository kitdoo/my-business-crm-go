<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()
const submitted = ref(false)

useSeoMeta({
  title: t('seo.contact.title'),
  description: t('seo.contact.description'),
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-2xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-4">{{ t('nav.kontakt') }}</h1>

    <div v-if="submitted" class="text-center py-16">
      <UIcon name="i-lucide-circle-check" class="w-14 h-14 text-brand-500 mx-auto mb-6" />
      <p class="text-lg">{{ t('contact.thanks') }}</p>
    </div>
    <ContactForm v-else @submitted="submitted = true" />
  </div>
</template>
