<script setup>
// Only used in SiteFooter's bottom bar now (moved off the TopBar).
const { t } = useI18n()
const config = useRuntimeConfig()
const adminUrl = config.public.adminUrl
</script>

<template>
  <a
    v-if="adminUrl"
    :href="adminUrl"
    target="_blank"
    rel="noopener"
    :aria-label="t('nav.adminLogin')"
    class="inline-flex items-center justify-center w-9 h-9 rounded-full text-white/60 hover:text-white hover:bg-white/10 transition-colors"
  >
    <UIcon name="i-lucide-circle-user" class="w-5 h-5" />
  </a>
</template>
