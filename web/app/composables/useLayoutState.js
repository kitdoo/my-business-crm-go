// Shared layout UI state (TD §8): whether the right-side menu is open as
// an overlay on < lg (TD §6).
export function useLayoutState() {
  const rightDrawerOpen = useState('layout-right-drawer-open', () => false)
  return { rightDrawerOpen }
}
