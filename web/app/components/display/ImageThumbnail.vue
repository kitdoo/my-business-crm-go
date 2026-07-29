<script setup>
// Small table-cell preview of the first image in an imageIds array (e.g.
// Product.imageIds) — same /api/images/:id route ProductImageUploader
// already uses to render previews, just read-only and tiny.
const props = defineProps({
  value: { type: Array, default: () => [] },
})
</script>

<template>
  <img v-if="value?.length" :src="`/api/images/${value[0]}?w=200`" class="size-8 rounded object-cover" alt="" />
  <div v-else class="size-8 rounded bg-neutral-100 flex items-center justify-center text-neutral-400">
    <UIcon name="i-lucide-image-off" class="size-4" />
  </div>
</template>
