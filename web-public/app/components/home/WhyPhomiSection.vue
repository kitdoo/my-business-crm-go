<script setup>
const { t } = useI18n()
</script>

<template>
  <!-- why_phomi.jpg is a finished infographic ("ZAŠTO PHOMI?" + callouts
       baked into the image) — full width, natural (unconstrained) height so
       nothing gets cropped. The catalog CTA lives on DezenStatsSection right
       above this, so no duplicate button here. -->
  <section class="w-full bg-[#1a1a1a] flex justify-center">
    <NuxtImg src="/images/why_phomi.jpg" :alt="t('home.whyPhomi.title')" loading="lazy" class="w-full lg:max-w-[70vw] h-auto" />
  </section>
</template>
