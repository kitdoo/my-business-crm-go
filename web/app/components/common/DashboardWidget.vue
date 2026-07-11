<script setup>
// One shared shell for every dashboard widget (TD §8.4/§9.3): title,
// loading, error+retry. Each widget loads its own data independently, so
// one broken report never blanks the rest of the dashboard.
defineProps({
  title: { type: String, required: true },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
})
const emit = defineEmits(['retry'])
</script>

<template>
  <div class="rounded-lg border border-neutral-200 p-4 space-y-3">
    <h3 class="font-semibold">{{ title }}</h3>
    <div v-if="loading" class="py-6 text-center text-neutral-500 text-sm">…</div>
    <div v-else-if="error" class="py-6 text-center text-sm space-y-2">
      <p class="text-red-600">{{ error }}</p>
      <UButton size="xs" variant="soft" @click="emit('retry')">{{ $t('common.retry') }}</UButton>
    </div>
    <slot v-else />
  </div>
</template>
