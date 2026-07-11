<script setup>
import { advantages } from '~/config/advantages.js'
const { t } = useI18n()
</script>

<template>
  <section class="mx-auto max-w-7xl px-4 lg:px-8 py-8 lg:py-12">
    <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide text-center mb-6">
      {{ t('home.advantages.title') }}
    </h2>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-8">
      <div v-for="a in advantages" :key="a.titleKey" class="text-center">
        <div class="mx-auto w-20 h-20 rounded-full bg-brand-500 flex items-center justify-center mb-5">
          <img :src="a.icon" alt="" class="w-9 h-9 [filter:brightness(0)_invert(1)]" />
        </div>
        <h3 class="text-lg lg:text-xl font-bold uppercase mb-3">{{ t(a.titleKey) }}</h3>
        <p class="text-sm text-black/60 leading-relaxed max-w-xs mx-auto">{{ t(a.textKey) }}</p>
      </div>
    </div>
  </section>
</template>
