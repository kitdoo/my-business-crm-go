<script setup>
// One badge component for every enum status in the project (TD §7) — each
// status maps to one of 5 shared color families instead of a bespoke set
// of colors per entity.
import { tokens } from '~/design/tokens.js'

const props = defineProps({
  status: { type: String, required: true }, // e.g. 'BRAND_STATUS_ACTIVE'
  map: { type: Object, required: true }, // STATUS_COLOR_MAP.<entity>
})

const family = computed(() => props.map[props.status] || 'inactive')
const colors = computed(() => tokens.color.status[family.value])

const { t } = useI18n()
const label = computed(() => t(`enums.status.${props.status}`, props.status))
</script>

<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
    :style="{ backgroundColor: colors.bg, color: colors.text }"
  >
    {{ label }}
  </span>
</template>
