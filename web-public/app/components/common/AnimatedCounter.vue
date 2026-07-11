<script setup>
// Counts up from 0 to `to` once the element enters the viewport — used by
// WorldPresenceSection and DezenStatsSection (TZ §4.3). Deliberately not a
// heavy animation library, just requestAnimationFrame over ~1.2s, plus a
// simple fade/slide-in gated on the same intersection trigger.
const props = defineProps({
  to: { type: Number, required: true },
  suffix: { type: String, default: '+' },
  label: { type: String, default: '' },
  durationMs: { type: Number, default: 2200 },
  // Set to false when the counter sits on a light/white section (label
  // color needs to invert — the number color is independent, see numberClass).
  dark: { type: Boolean, default: true },
  numberClass: { type: String, default: 'text-brand-500' },
  numberSizeClass: { type: String, default: '!text-4xl lg:!text-5xl' },
})

const current = ref(0)
const visible = ref(false)
const el = ref(null)
let started = false

function animate() {
  if (started) return
  started = true
  visible.value = true
  const start = performance.now()
  function tick(now) {
    const progress = Math.min(1, (now - start) / props.durationMs)
    current.value = Math.round(props.to * (1 - Math.pow(1 - progress, 3)))
    if (progress < 1) requestAnimationFrame(tick)
  }
  requestAnimationFrame(tick)
}

onMounted(() => {
  if (typeof IntersectionObserver === 'undefined') {
    animate()
    return
  }
  const observer = new IntersectionObserver((entries) => {
    if (entries.some((e) => e.isIntersecting)) animate()
  }, { threshold: 0.3 })
  if (el.value) observer.observe(el.value)
  onUnmounted(() => observer.disconnect())
})
</script>

<template>
  <div
    ref="el"
    class="text-center transition-all duration-700 ease-out"
    :class="visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
  >
    <div class="text-hero font-bold" :class="[numberSizeClass, numberClass]">{{ current }}{{ suffix }}</div>
    <div v-if="label" class="mt-1 text-sm uppercase tracking-wide" :class="dark ? 'text-white/70' : 'text-black/60'">
      {{ label }}
    </div>
  </div>
</template>
