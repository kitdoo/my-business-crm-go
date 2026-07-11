<script setup>
// Period presets (TD §8.4/§9.3) — reused by every report widget on the
// dashboard, all sharing one selected range. Custom/arbitrary range
// isn't implemented yet, only the three presets.
const props = defineProps({
  modelValue: { type: Object, default: () => ({}) }, // { from, to } unix seconds
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

const PRESETS = [
  { key: 'today', label: 'periodFilter.today', days: 0 },
  { key: '7d', label: 'periodFilter.7d', days: 7 },
  { key: '30d', label: 'periodFilter.30d', days: 30 },
]
const active = ref('7d')

function apply(preset) {
  active.value = preset.key
  const to = Math.floor(Date.now() / 1000)
  const from =
    preset.days === 0 ? Math.floor(new Date(new Date().setHours(0, 0, 0, 0)).getTime() / 1000) : to - preset.days * 86400
  emit('update:modelValue', { from, to })
}

onMounted(() => apply(PRESETS[1]))
</script>

<template>
  <div class="flex gap-2">
    <UButton
      v-for="preset in PRESETS"
      :key="preset.key"
      :variant="active === preset.key ? 'solid' : 'soft'"
      size="sm"
      @click="apply(preset)"
    >
      {{ t(preset.label) }}
    </UButton>
  </div>
</template>
