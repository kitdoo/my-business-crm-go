// Flips `visible` to true once when `el` enters the viewport — same pattern
// as AnimatedCounter.vue, reused here to gate the infographic line-draw /
// badge / label reveal animation.
export function useRevealOnScroll(threshold = 0.25) {
  const el = ref(null)
  const visible = ref(false)

  onMounted(() => {
    if (typeof IntersectionObserver === 'undefined') {
      visible.value = true
      return
    }
    const observer = new IntersectionObserver((entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        visible.value = true
        observer.disconnect()
      }
    }, { threshold })
    if (el.value) observer.observe(el.value)
    onUnmounted(() => observer.disconnect())
  })

  return { el, visible }
}
