<script setup>
// Compact KPI figure (dataviz skill's "stat tile" contract): label +
// value only, no delta/trend — the dashboard has no prior-period figure
// to compare against yet. Loading/error mirror DashboardWidget's shell
// so a broken report still fails independently, just in a smaller box.
defineProps({
  label: { type: String, required: true },
  loading: { type: Boolean, default: false },
  error: { type: String, default: '' },
})
</script>

<template>
  <div class="rounded-lg border border-neutral-200 p-3 min-w-[9rem]">
    <div class="text-xs text-neutral-500">{{ label }}</div>
    <div v-if="loading" class="text-lg text-neutral-400">…</div>
    <div v-else-if="error" class="text-xs text-red-600">{{ error }}</div>
    <div v-else class="text-2xl font-semibold"><slot /></div>
  </div>
</template>
