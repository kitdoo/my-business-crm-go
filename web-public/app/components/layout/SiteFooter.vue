<script setup>
const { t } = useI18n()
const localePath = useLocalePath()
const year = new Date().getFullYear()
</script>

<template>
  <footer class="bg-[#333333] text-white">
    <div class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 py-12 lg:py-16 grid grid-cols-1 md:grid-cols-4 gap-10">
      <div>
        <NuxtLink :to="localePath('/')" class="inline-flex flex-col items-center gap-1 mb-4">
          <img src="/images/logos/logo_blue.png" alt="PHOMI SRBIJA" class="h-[72px] w-auto object-contain" />
          <span class="tracking-[0.2em] text-sm">
            <span class="font-bold text-brand-500">PHOMI</span> <span class="font-light text-white">SRBIJA</span>
          </span>
        </NuxtLink>
        <div class="flex items-center gap-4">
          <a href="https://www.instagram.com/phomi_serbia/" target="_blank" rel="noopener" aria-label="Instagram"><UIcon name="i-lucide-instagram" class="w-5 h-5 text-white/70 hover:text-white" /></a>
          <a href="mailto:serbia@phomi.info" aria-label="Email"><UIcon name="i-lucide-mail" class="w-5 h-5 text-white/70 hover:text-white" /></a>
        </div>
      </div>

      <div>
        <h3 class="text-base font-semibold uppercase text-brand-300 mb-3">{{ t('footer.about') }}</h3>
        <ul class="space-y-2 text-base text-white/80">
          <li><a href="https://phomi.com/" target="_blank" rel="noopener">{{ t('footer.aboutPhomi') }}</a></li>
          <li><a href="https://tom.archi/" target="_blank" rel="noopener">{{ t('footer.aboutTomStudio') }}</a></li>
        </ul>
      </div>

      <div>
        <h3 class="text-base font-semibold uppercase text-brand-300 mb-3">{{ t('nav.katalog') }}</h3>
        <ul class="space-y-2 text-base text-white/80">
          <li><NuxtLink :to="localePath('/katalog')">{{ t('footer.ekoPaneli') }}</NuxtLink></li>
          <li><NuxtLink :to="localePath('/katalog')">{{ t('footer.solarniPaneli') }}</NuxtLink></li>
          <li><NuxtLink :to="localePath('/katalog')">{{ t('footer.displejPaneli') }}</NuxtLink></li>
        </ul>
      </div>

      <div>
        <h3 class="text-base font-semibold uppercase text-brand-300 mb-3">{{ t('footer.usefulLinks') }}</h3>
        <ul class="space-y-2 text-base text-white/80">
          <li><NuxtLink :to="localePath('/postani-diler')">{{ t('nav.postaniDiler') }}</NuxtLink></li>
          <li><NuxtLink :to="localePath('/kontakt')">{{ t('nav.kontakt') }}</NuxtLink></li>
          <li><NuxtLink :to="localePath('/projekti')">{{ t('nav.zavrseniProjekti') }}</NuxtLink></li>
        </ul>
      </div>
    </div>

    <div class="border-t border-white/10">
      <div class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 py-2 flex items-center justify-between">
        <AdminLoginLink />
        <p class="text-sm text-white/60">© {{ year }} Tom Studio 021 doo x Kitdoo</p>
      </div>
    </div>
  </footer>
</template>
