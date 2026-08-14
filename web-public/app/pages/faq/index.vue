<script setup>
const { t } = useI18n()
const localeHead = useSeoHead()
const localePath = useLocalePath()

useSeoMeta({
  title: t('seo.faq.title'),
  description: t('seo.faq.description'),
  ogTitle: t('seo.faq.title'),
  ogDescription: t('seo.faq.description'),
  ogImage: useAbsoluteUrl('/images/social-share.jpg'),
})

// Fixed q1..q7/a1..a7 keys (see locale files) rather than a translated
// array — the rest of the project's i18n content is flat keyed strings,
// not arrays (vue-i18n needs tm(), not t(), to read arrays back out, and
// nothing else here uses that).
const QUESTION_COUNT = 7
const items = computed(() => Array.from({ length: QUESTION_COUNT }, (_, i) => ({
  q: t(`faq.q${i + 1}`),
  a: t(`faq.a${i + 1}`),
})))

// FAQPage structured data (schema.org) — read by Google's FAQ rich result
// and AI answer engines, same rationale as the Product JSON-LD on
// katalog/[sku].vue.
const faqJsonLd = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'FAQPage',
  mainEntity: items.value.map((item) => ({
    '@type': 'Question',
    name: item.q,
    acceptedAnswer: { '@type': 'Answer', text: item.a },
  })),
}))
useHead(() => ({
  link: localeHead.value.link,
  meta: localeHead.value.meta,
  script: [{ type: 'application/ld+json', innerHTML: JSON.stringify(faqJsonLd.value) }],
}))
</script>

<template>
  <div class="mx-auto max-w-3xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-8">{{ t('faq.title') }}</h1>

    <div class="divide-y divide-black/10 border-t border-b border-black/10">
      <details v-for="(item, i) in items" :key="i" class="group py-4">
        <summary class="flex items-center justify-between gap-4 cursor-pointer list-none font-medium text-black/90">
          {{ item.q }}
          <UIcon name="i-lucide-chevron-down" class="w-5 h-5 shrink-0 text-black/40 transition-transform group-open:rotate-180" />
        </summary>
        <p class="text-black/70 leading-relaxed mt-3">{{ item.a }}</p>
      </details>
    </div>

    <p class="text-black/70 leading-relaxed mt-10 text-center">
      {{ t('faq.cta') }}
      <NuxtLink :to="localePath('/kontakt')" class="text-brand-700 font-medium underline underline-offset-4">
        {{ t('faq.ctaLink') }}
      </NuxtLink>
    </p>
  </div>
</template>
