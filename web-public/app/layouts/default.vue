<script setup>
const { isCompact } = useHeaderCompact()
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <TopBar />
    <!--
      TopBar is `fixed` (see TopBar.vue) so the hero photo on the home page
      can sit full-bleed behind the logo/menu (transparent header over the
      image). That takes TopBar out of document flow, so every other page
      needs this top padding to not render underneath it; HeroSection
      cancels it back out with a matching negative margin — see below. The
      padding shrinks together with TopBar's own height (useHeaderCompact)
      so scrolling doesn't leave a growing gap under the now-thinner bar.
    -->
    <main class="flex-1 transition-[padding] duration-300" :class="isCompact ? 'pt-16 lg:pt-20' : 'pt-28 lg:pt-32'">
      <slot />
    </main>
    <SiteFooter />
  </div>
</template>
