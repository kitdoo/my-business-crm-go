<script setup>
const { t } = useI18n()
const localePath = useLocalePath()
</script>

<template>
  <section class="relative h-screen min-h-[560px] flex items-center justify-start overflow-hidden">
    <!-- map.jpg is pre-cropped so the arrow hub sits at ~72%/43% of the
         image itself (right side, clear of the text panel on the left) —
         object-position keeps that exact point anchored at the same spot
         in the viewport regardless of aspect ratio, unlike a CSS scale()
         zoom (which re-zooms from center and drifts the hub off-screen on
         tall/narrow viewports). -->
    <NuxtImg src="/images/background/map.jpg" alt="" loading="lazy" class="absolute inset-0 w-full h-full object-cover object-[72%_43%]" />
    <div class="relative z-10 ml-6 md:ml-16 lg:ml-24 max-w-xl rounded-sm bg-[#333333]/30 px-8 py-10 lg:px-12 lg:py-14 text-white backdrop-blur-sm">
      <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-6">{{ t('home.becomePartner.title') }}</h2>
      <p class="text-white/90 mb-2">{{ t('home.becomePartner.text') }}</p>
      <p class="text-white/90 mb-8">{{ t('home.becomePartner.text2') }}</p>
      <UButton :to="localePath('/postani-diler')" variant="cta-white" size="xl" class="px-10 py-4 text-base">
        {{ t('home.becomePartner.cta') }}
      </UButton>
    </div>
  </section>
</template>
