/** Shared by TopBar and the default layout so the fixed header's shrink-on-
 * scroll state and the in-flow spacer that reserves room for it never
 * disagree — see TopBar.vue and layouts/default.vue. */
export function useHeaderCompact() {
  const route = useRoute()
  const localePath = useLocalePath()
  const { scrolled } = useScrollHeader()

  const isHome = computed(() => route.path === localePath('/'))
  // Non-home pages always show the solid header, and scrolling down on any
  // page means the visitor is reading content, not looking at the logo —
  // in both cases shrink the header so the bar stays a thin strip.
  const isCompact = computed(() => !isHome.value || scrolled.value)

  return { scrolled, isHome, isCompact }
}
