<script setup>
const { t } = useI18n()
</script>

<template>
  <!-- how_to_install.jpg is a finished infographic ("KAKO SE UGRAĐUJE
       PHOMI?" + install-layer callouts baked into the image) — full width,
       natural (unconstrained) height so nothing gets cropped. -->
  <section class="w-full bg-[#1a1a1a] flex justify-center">
    <NuxtImg src="/images/how_to_install.jpg" :alt="t('home.advantages.install.title')" loading="lazy" class="w-full lg:max-w-[70vw] h-auto" />
  </section>
</template>
