<script setup>
// Composes a project's real material photos into the layout its tile count
// implies — 1 photo stays full-width, 2 sit side by side, 3 is a 2-up row
// with a full-width tile underneath, 4 is a 2x2 grid — each tile carrying
// its own name/color/size label baked as an HTML overlay (crisp at any
// size, no raster text-on-image step needed).
//
// Sizing: on lg+ the parent puts this next to the project's ImageSlider in
// a stretched flex row, so it fills that full height via flex-1 (passed in
// as a class from the caller) + auto-rows-fr, reaching the same bottom
// edge as the gallery image instead of leaving whitespace under shorter
// content. Below lg the gallery image stacks above this block at w-full
// aspect-[4/3]; aspect-[8/3] here makes the whole collage exactly half
// that image's height, whatever the tile count.
defineProps({
  materials: { type: Array, required: true },
})
</script>

<template>
  <div
    class="grid gap-3 auto-rows-fr aspect-[8/3] lg:aspect-auto"
    :class="materials.length === 1 ? 'grid-cols-1' : 'grid-cols-2'"
  >
    <div
      v-for="(m, i) in materials"
      :key="m.image"
      class="relative overflow-hidden rounded-sm bg-gray-100 h-full"
      :class="materials.length === 3 && i === 2 ? 'col-span-2' : ''"
    >
      <NuxtImg :src="m.image" :alt="`${m.name} ${m.color}`" loading="lazy" class="absolute inset-0 w-full h-full object-cover" />
      <div class="absolute inset-x-0 bottom-0 px-4 py-3 bg-gradient-to-t from-black/75 via-black/30 to-transparent text-white">
        <p class="text-lg sm:text-xl font-extrabold uppercase tracking-wide leading-tight">{{ m.name }}</p>
        <p class="text-sm sm:text-base font-light uppercase tracking-[0.15em] leading-tight">
          {{ m.color }}<span v-if="m.size" class="text-white/70 tracking-normal font-light"> &nbsp;{{ m.size }}</span>
        </p>
      </div>
    </div>
  </div>
</template>
