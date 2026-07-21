<script setup>
// A single callout line (2-3 point polyline) that "draws itself in" via
// stroke-dasharray/stroke-dashoffset once the parent section becomes
// visible (TZ §9 — line draw, then badge, then label).
const props = defineProps({
  points: { type: Array, required: true }, // [[x,y], [x,y], ...] in viewBox units
  visible: { type: Boolean, default: false },
  delayMs: { type: Number, default: 0 },
})

// points arrives with its final (clearance-adjusted) values only after
// InfographicPanel's ResizeObserver measures the real badge size post-mount
// — both of these must stay reactive to that later update, not just the
// initial mount, or the polyline silently keeps rendering the pre-measurement
// coordinates forever.
const pointsAttr = computed(() => props.points.map(([x, y]) => `${x},${y}`).join(' '))
const line = ref(null)
const length = ref(0)

async function updateLength() {
  await nextTick()
  if (line.value) length.value = line.value.getTotalLength()
}

onMounted(updateLength)
watch(() => props.points, updateLength, { deep: true })
</script>

<template>
  <polyline
    ref="line"
    :points.attr="pointsAttr"
    fill="none"
    stroke="#ffffff"
    stroke-opacity="0.8"
    stroke-width="1.5"
    :style="{
      strokeDasharray: length,
      strokeDashoffset: visible ? 0 : length,
      transition: `stroke-dashoffset 700ms ease-out ${delayMs}ms`,
    }"
  />
</template>
