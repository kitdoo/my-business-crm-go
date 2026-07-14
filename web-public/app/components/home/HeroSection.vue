<script setup>
const { t } = useI18n()
</script>

<template>
  <!-- -mt cancels out the layout's pt-28/lg:pt-32 (added so the fixed
       TopBar doesn't cover normal pages) so this photo fills the full
       viewport from y=0, right behind the transparent TopBar. -->
  <section class="relative -mt-28 lg:-mt-32 h-screen min-h-[560px] flex items-center justify-end">
    <NuxtImg
      src="/images/background/top.jpeg"
      alt=""
      class="absolute inset-0 w-full h-full object-cover"
      preload
      fetchpriority="high"
    />
    <div class="absolute inset-0 bg-black/30" />
    <div class="relative z-10 mx-auto max-w-7xl 2xl:max-w-[1800px] w-full px-4 lg:px-8 -mt-28 lg:-mt-40 text-white text-right">
      <h1 class="text-hero max-w-3xl ml-auto">{{ t('home.hero.title') }}</h1>
      <p class="mt-4 max-w-xl ml-auto text-lg text-white/90">{{ t('home.hero.subtitle') }}</p>
    </div>
    <UIcon
      name="i-lucide-chevron-down"
      class="absolute z-10 bottom-8 right-1/4 w-6 h-6 text-white/80 animate-bounce"
    />
  </section>
</template>
