<script setup>
// Bare icon used by InfographicPanel callouts (no circle backdrop, per user
// feedback). Icons in public/images/icons/animation_* come in mixed source
// colors (some black strokes, some white) — forcing them white via CSS
// filter keeps every icon visually consistent regardless of the source
// PNG's own color.
defineProps({
  icon: { type: String, required: true },
})
</script>

<template>
  <div
    class="flex items-center justify-center shrink-0"
    style="width: clamp(30px, 5.5vw, 64px); height: clamp(30px, 5.5vw, 64px);"
  >
    <img :src="icon" alt="" class="w-full h-full object-contain [filter:brightness(0)_invert(1)]" />
  </div>
</template>
