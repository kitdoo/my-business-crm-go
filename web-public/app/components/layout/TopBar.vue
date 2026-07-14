<script setup>
const { t, locale, locales, setLocale } = useI18n()
const localePath = useLocalePath()
const { isHome, isCompact, scrolled } = useHeaderCompact()
const { count: cartCount, open: openCart } = useCart()
const mobileOpen = ref(false)

const navItems = computed(() => [
  { label: t('nav.katalog'), to: localePath('/katalog') },
  { label: t('nav.zavrseniProjekti'), to: localePath('/projekti') },
  { label: t('nav.kontakt'), to: localePath('/kontakt') },
  { label: t('nav.postaniDiler'), to: localePath('/postani-diler') },
])

const availableLocales = computed(() => locales.value)

// Only the home page has a full-bleed hero photo behind the header, so
// only there is a transparent-then-brand-on-scroll header appropriate —
// every other page gets a plain solid black header (item 13), since a
// transparent one over plain white content was unreadable.

const langMenuItems = computed(() => [
  availableLocales.value.map((l) => ({
    label: l.code.toUpperCase(),
    class: 'text-black data-[highlighted]:text-black data-[highlighted]:bg-black/5',
    onSelect: () => setLocale(l.code),
  })),
])

function closeMobile() {
  mobileOpen.value = false
}
</script>

<template>
  <header
    class="fixed inset-x-0 top-0 z-50 text-white transition-colors duration-300"
    :class="!isHome ? 'bg-black' : (scrolled ? 'bg-brand-500 shadow-sm' : 'bg-transparent')"
  >
    <div
      class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 flex items-center justify-between transition-[min-height] duration-300"
      :class="isCompact ? 'min-h-16 lg:min-h-20' : 'min-h-28 lg:min-h-32'"
    >
      <NuxtLink :to="localePath('/')" class="-ml-2 lg:-ml-4 flex flex-col items-center transition-all duration-300" :class="isCompact ? 'gap-0 pt-3' : 'gap-1 pt-9 lg:pt-11'">
        <img
          src="/images/logos/logo_white.png"
          alt="PHOMI SRBIJA"
          class="w-auto object-contain transition-[height] duration-300"
          :class="isCompact ? 'h-8 lg:h-10' : 'h-[72px] lg:h-[88px]'"
        />
        <span class="tracking-[0.2em] text-white transition-[font-size] duration-300" :class="isCompact ? 'text-[10px] lg:text-xs' : 'text-sm lg:text-base'">
          <span class="font-bold">PHOMI</span> <span class="font-light">SRBIJA</span>
        </span>
      </NuxtLink>

      <div class="hidden lg:flex items-center gap-6">
        <nav class="flex items-center gap-8 text-sm font-medium uppercase tracking-wide">
          <NuxtLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="hover:text-brand-500 transition-colors"
            active-class="text-brand-500"
          >
            {{ item.label }}
          </NuxtLink>
        </nav>

        <div class="w-px h-5 bg-white/30" />

        <UDropdownMenu :items="langMenuItems" :ui="{ content: 'bg-white text-black' }">
          <button class="text-sm font-medium uppercase" :aria-label="t('nav.langSwitch')">
            {{ locale }}
          </button>
        </UDropdownMenu>

        <button class="relative" :aria-label="t('cart.title')" @click="openCart">
          <UIcon name="i-lucide-shopping-cart" class="w-6 h-6" />
          <span
            v-if="cartCount"
            class="absolute -top-2 -right-2 min-w-4 h-4 px-1 rounded-full bg-brand-500 text-white text-[10px] font-bold flex items-center justify-center"
          >
            {{ cartCount }}
          </span>
        </button>
      </div>

      <div class="flex items-center gap-4 lg:hidden">
        <button class="relative" :aria-label="t('cart.title')" @click="openCart">
          <UIcon name="i-lucide-shopping-cart" class="w-6 h-6" />
          <span
            v-if="cartCount"
            class="absolute -top-2 -right-2 min-w-4 h-4 px-1 rounded-full bg-brand-500 text-white text-[10px] font-bold flex items-center justify-center"
          >
            {{ cartCount }}
          </span>
        </button>
        <button :aria-label="t('nav.menu')" @click="mobileOpen = true">
          <UIcon name="i-lucide-menu" class="w-7 h-7" />
        </button>
      </div>
    </div>

    <!-- Mobile overlay menu — TZ §4.1. Full-screen only on real phone
         widths; from `sm:` up it's a right-anchored drawer so it doesn't
         look like an empty full-screen blank on tablets. -->
    <div
      v-if="mobileOpen"
      class="fixed inset-0 sm:inset-y-0 sm:left-auto sm:right-0 sm:w-96 sm:max-w-[85vw] z-[60] bg-[#333333] text-white flex flex-col shadow-2xl"
    >
      <div class="flex items-center justify-between px-4 h-16">
        <span class="font-bold">PHOMI SRBIJA</span>
        <button :aria-label="t('nav.close')" @click="closeMobile">
          <UIcon name="i-lucide-x" class="w-7 h-7" />
        </button>
      </div>
      <nav class="flex-1 flex flex-col items-start justify-center gap-6 px-8 text-2xl font-semibold uppercase">
        <NuxtLink v-for="item in navItems" :key="item.to" :to="item.to" @click="closeMobile">
          {{ item.label }}
        </NuxtLink>
      </nav>
      <div class="flex items-center gap-4 px-8 py-6 border-t border-white/10">
        <button
          v-for="l in availableLocales"
          :key="l.code"
          class="text-sm font-medium uppercase"
          :class="{ 'text-brand-500': l.code === locale }"
          @click="setLocale(l.code); closeMobile()"
        >
          {{ l.code }}
        </button>
      </div>
    </div>

    <CartDrawer />
  </header>
</template>
