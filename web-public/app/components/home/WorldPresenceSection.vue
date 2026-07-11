<script setup>
const { t } = useI18n()

const stats = [
  { to: 500, label: 'home.worldPresence.projects' },
  { to: 400, label: 'home.worldPresence.dealers' },
  { to: 20, label: 'home.worldPresence.countries' },
  { to: 30, label: 'home.worldPresence.patents' },
]
</script>

<template>
  <section class="bg-white text-black py-16 lg:py-20">
    <div class="mx-auto max-w-5xl px-4 lg:px-8 text-center">
      <h2 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-3">
        {{ t('home.worldPresence.title') }}
      </h2>
      <p class="text-lg text-brand-600 font-medium mb-5">{{ t('home.worldPresence.subtitle') }}</p>
      <p class="text-black/70 leading-relaxed mb-4">{{ t('home.worldPresence.intro1') }}</p>
      <p class="text-black/70 leading-relaxed">{{ t('home.worldPresence.intro2') }}</p>
    </div>

    <div class="mx-auto max-w-7xl px-6 lg:px-12 mt-12">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-x-6 gap-y-10 mb-12 max-w-4xl mx-auto">
        <AnimatedCounter
          v-for="s in stats"
          :key="s.label"
          :to="s.to"
          :dark="false"
          number-class="text-black"
          number-size-class="!text-5xl lg:!text-6xl"
          :label="t(s.label)"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-10 max-w-5xl mx-auto text-center md:text-left">
        <div>
          <h3 class="text-lg font-bold uppercase mb-3">{{ t('home.worldPresence.growth.title') }}</h3>
          <p class="text-black/70 leading-relaxed">{{ t('home.worldPresence.growth.text') }}</p>
        </div>
        <div>
          <h3 class="text-lg font-bold uppercase mb-3">{{ t('home.worldPresence.scale.title') }}</h3>
          <p class="text-black/70 leading-relaxed">{{ t('home.worldPresence.scale.text') }}</p>
        </div>
      </div>
    </div>
  </section>
</template>
