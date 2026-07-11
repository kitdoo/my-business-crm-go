<script setup>
const { t } = useI18n()
</script>

<template>
  <section class="mx-auto max-w-4xl px-4 lg:px-8 py-8 lg:py-12 text-center">
    <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide">{{ t('home.about.title') }}</h2>
    <p class="mt-6 text-lg lg:text-xl font-medium text-black/80 leading-relaxed">{{ t('home.about.text') }}</p>
    <p class="mt-5 text-black/60 leading-relaxed">{{ t('home.about.text2') }}</p>
    <p class="mt-4 text-black/60 leading-relaxed">{{ t('home.about.text3') }}</p>
  </section>
</template>
