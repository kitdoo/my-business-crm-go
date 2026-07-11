<script setup>
// Adaptive form layout (TD §6/§9.3): one column below `md`, two columns
// at `md` and up.
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <slot />
  </div>
</template>
