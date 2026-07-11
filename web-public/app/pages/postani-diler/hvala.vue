<script setup>
const { t } = useI18n()
const localePath = useLocalePath()
useSeoMeta({ title: t('seo.dealerThanks.title') })
</script>

<template>
  <div class="mx-auto max-w-xl px-4 lg:px-8 py-24 text-center">
    <UIcon name="i-lucide-circle-check" class="w-14 h-14 text-brand-500 mx-auto mb-6" />
    <h1 class="text-2xl font-bold mb-4">{{ t('dealer.thanksTitle') }}</h1>
    <p class="text-black/70 mb-10">{{ t('dealer.thanksText') }}</p>
    <UButton :to="localePath('/')" variant="cta-outline" size="lg">{{ t('common.backHome') }}</UButton>
  </div>
</template>
