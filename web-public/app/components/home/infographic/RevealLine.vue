<script setup>
// A single callout line (2-3 point polyline) that "draws itself in" via
// stroke-dasharray/stroke-dashoffset once the parent section becomes
// visible (TZ §9 — line draw, then badge, then label).
const props = defineProps({
  points: { type: Array, required: true }, // [[x,y], [x,y], ...] in viewBox units
  visible: { type: Boolean, default: false },
  delayMs: { type: Number, default: 0 },
})

const pointsAttr = props.points.map(([x, y]) => `${x},${y}`).join(' ')
const line = ref(null)
const length = ref(0)

onMounted(() => {
  if (line.value) length.value = line.value.getTotalLength()
})
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
