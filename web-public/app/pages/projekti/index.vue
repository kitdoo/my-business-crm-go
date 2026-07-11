<script setup>
import { projects } from '~/config/projects.js'

const { t } = useI18n()
const localeHead = useLocaleHead()

useSeoMeta({
  title: t('seo.projects.title'),
  description: t('seo.projects.description'),
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-10">{{ t('nav.zavrseniProjekti') }}</h1>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
      <figure v-for="p in projects" :key="p.id" class="rounded-lg overflow-hidden">
        <NuxtImg :src="p.image" :alt="t(p.titleKey)" loading="lazy" class="w-full h-64 object-cover" />
        <figcaption class="mt-3 text-sm text-black/70">{{ t(p.titleKey) }}</figcaption>
      </figure>
    </div>
  </div>
</template>
