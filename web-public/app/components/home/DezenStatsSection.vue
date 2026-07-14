<script setup>
const { t } = useI18n()
const localePath = useLocalePath()

const stats = [
  { to: 640, label: 'home.dezenStats.stone' },
  { to: 30, label: 'home.dezenStats.wood' },
  { to: 56, label: 'home.dezenStats.woven' },
  { to: 42, label: 'home.dezenStats.brick' },
]
</script>

<template>
  <section class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-6 lg:px-12 py-8 lg:py-12 text-center">
    <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-6">{{ t('home.dezenStats.title') }}</h2>
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-x-6 gap-y-10 mb-8 max-w-4xl mx-auto">
      <AnimatedCounter
        v-for="s in stats"
        :key="s.label"
        :to="s.to"
        suffix=""
        :dark="false"
        number-class="text-black"
        number-size-class="!text-5xl lg:!text-6xl"
        :label="t(s.label)"
      />
    </div>
    <UButton :to="localePath('/katalog')" variant="cta-outline" size="xl" class="px-10 py-4 text-base">
      {{ t('home.whyPhomi.cta') }}
    </UButton>
  </section>
</template>
