// Shared radial-gradient background for the infographic sections (see
// InfographicPanel.vue) and their full-width section wrappers, so the
// backdrop can extend edge-to-edge while the card content stays centered.
export function useInfographicGradient(origin) {
  return computed(() => {
    const pos = origin.value === 'tr' ? '100% 0%' : '100% 100%'
    return {
      background: `radial-gradient(130% 130% at ${pos}, #3f4247 0%, #232529 35%, #17181b 65%, #0b0c0d 100%)`,
    }
  })
}
