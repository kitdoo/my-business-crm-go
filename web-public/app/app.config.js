import { tokens } from '~/design/tokens.js'

export default defineAppConfig({
  ui: {
    colors: {
      primary: 'brand',
      neutral: 'neutral',
    },
    button: {
      variants: {
        // Outline CTA per TZ §5 — turquoise border, transparent/dark
        // background, turquoise text. Used as <UButton variant="cta-outline">.
        variant: {
          'cta-outline': 'text-brand-500 ring ring-inset ring-brand-500 bg-transparent hover:bg-brand-500/10',
          // Solid white CTA for use over photo backgrounds where the
          // outline variant doesn't read well (e.g. BecomePartnerSection).
          'cta-white': 'text-[#333333] bg-white ring ring-inset ring-white hover:bg-white/90',
        },
      },
    },
  },
  brand: tokens,
})
