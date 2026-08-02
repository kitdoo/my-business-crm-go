<script setup>
import { projects } from '~/config/projects.js'

const { t } = useI18n()
const localeHead = useLocaleHead()

useSeoMeta({
  title: t('seo.projects.title'),
  description: t('seo.projects.description'),
  ogTitle: t('seo.projects.title'),
  ogDescription: t('seo.projects.description'),
  ogImage: useAbsoluteUrl('/images/social-share.jpg'),
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))
</script>

<template>
  <div class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-10">{{ t('nav.zavrseniProjekti') }}</h1>

    <div class="flex flex-col gap-16 lg:gap-20">
      <article v-for="p in projects" :key="p.id" class="w-full flex flex-col lg:flex-row gap-6">
        <figure class="w-full lg:w-[60%] flex-shrink-0">
          <ImageSlider :images="p.images" :alt="t(p.titleKey)" :caption="t(p.titleKey)" class="aspect-[4/3]" />
        </figure>

        <div class="w-full lg:w-[40%] flex flex-col gap-3">
          <p class="text-xs uppercase tracking-wide text-black/50">{{ t('projects.usedMaterials') }}</p>
          <MaterialsCollage :materials="p.materials" class="lg:flex-1 lg:min-h-0" />
        </div>
      </article>
    </div>
  </div>
</template>
