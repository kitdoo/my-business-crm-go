// Shared layout UI state (TD §8): whether the left-side menu is open as
// an overlay on < lg (TD §6).
export function useLayoutState() {
  const menuOpen = useState('layout-menu-open', () => false)
  return { menuOpen }
}
