<script setup>
import { projects } from '~/config/projects.js'
const { t } = useI18n()
const localePath = useLocalePath()
const preview = projects.slice(0, 2)
</script>

<template>
  <section class="mx-auto max-w-7xl px-4 lg:px-8 py-8 lg:py-12">
    <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide text-center mb-5">
      {{ t('home.projectsPreview.title') }}
    </h2>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-3xl mx-auto">
      <NuxtLink
        v-for="p in preview"
        :key="p.id"
        :to="localePath('/projekti')"
        class="block rounded-lg overflow-hidden group"
      >
        <NuxtImg
          :src="p.image"
          :alt="t(p.titleKey)"
          loading="lazy"
          class="w-full aspect-square object-cover group-hover:scale-105 transition-transform duration-300"
        />
      </NuxtLink>
    </div>
    <div class="text-center mt-8">
      <UButton :to="localePath('/projekti')" variant="cta-outline" size="xl" class="px-10 py-4 text-base">
        {{ t('home.projectsPreview.cta') }}
      </UButton>
    </div>
  </section>
</template>
