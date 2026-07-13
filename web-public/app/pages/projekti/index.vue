<script setup>
import { projects } from '~/config/projects.js'

const { t } = useI18n()
const localeHead = useLocaleHead()
const localePath = useLocalePath()

useSeoMeta({
  title: t('seo.projects.title'),
  description: t('seo.projects.description'),
  ogTitle: t('seo.projects.title'),
  ogDescription: t('seo.projects.description'),
  ogImage: '/images/project_1.jpg',
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))

// Combined top-to-bottom stack: used materials first, recommended/similar
// last, each tile carrying its own label so the split stays visible
// without needing a separate 3-across grid per group.
function materialTiles(p) {
  return [
    ...p.usedMaterials.map((m) => ({ ...m, kind: 'used' })),
    ...p.recommendedMaterials.map((m) => ({ ...m, kind: 'recommended' })),
  ]
}
</script>

<template>
  <div class="mx-auto max-w-7xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-10">{{ t('nav.zavrseniProjekti') }}</h1>

    <div class="flex flex-col gap-16 lg:gap-20">
      <article v-for="p in projects" :key="p.id" class="w-full flex flex-col lg:flex-row gap-6">
        <figure class="w-full lg:w-[60%] flex-shrink-0">
          <ImageSlider :images="p.images" :alt="t(p.titleKey)" :caption="t(p.titleKey)" class="aspect-[4/3]" />
        </figure>

        <div class="w-full lg:w-[40%] flex flex-col gap-3">
          <NuxtLink
            v-for="m in materialTiles(p)"
            :key="m.id"
            :to="localePath('/katalog')"
            class="flex items-center gap-3 rounded-sm border border-black/10 overflow-hidden group"
          >
            <div class="w-20 h-20 flex-shrink-0 bg-gray-50">
              <NuxtImg :src="m.image" alt="" loading="lazy" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
            </div>
            <span class="text-xs uppercase tracking-wide text-black/50 pr-3">
              {{ m.kind === 'used' ? t('projects.usedMaterials') : t('projects.recommendedMaterials') }}
            </span>
          </NuxtLink>
        </div>
      </article>
    </div>
  </div>
</template>
