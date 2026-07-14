<script setup>
// Gradient origin 'br' matches the source graphic's bright corner so the
// full-width backdrop blends seamlessly with the centered card.
const gradientStyle = useInfographicGradient(computed(() => 'br'))
</script>

<template>
  <section class="w-full flex justify-center py-8 lg:py-12 px-4" :style="gradientStyle">
    <div class="w-full max-w-4xl">
      <WhyPhomiInfographic />
    </div>
  </section>
</template>
