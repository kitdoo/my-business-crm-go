/** Sticky TopBar switches from transparent-over-hero to solid on scroll — TZ §4.1. */
export function useScrollHeader(thresholdPx = 24) {
  const scrolled = ref(false)

  function onScroll() {
    scrolled.value = window.scrollY > thresholdPx
  }

  onMounted(() => {
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
  })
  onUnmounted(() => {
    window.removeEventListener('scroll', onScroll)
  })

  return { scrolled }
}
