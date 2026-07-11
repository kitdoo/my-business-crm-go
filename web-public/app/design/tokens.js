/**
 * Copied from web/app/design/tokens.js (admin panel) and extended with
 * typography.hero for the public site (TZ §5). Keep both files in sync by
 * hand until this repo moves to a shared workspace package (TZ §1.3).
 */
export const tokens = {
  color: {
    brand: {
      50: '#f4f9fa',
      100: '#e8f3f5',
      200: '#d0e6ea',
      300: '#b7dbe1',
      400: '#9ed0d8',
      500: '#85c5cf',
      600: '#6ba9b2',
      700: '#558790',
      800: '#3f656c',
      900: '#2a4448',
    },
    text: {
      primary: '#000000',
      secondary: '#525252',
      inverse: '#ffffff',
    },
    surface: {
      base: '#ffffff',
      subtle: '#f9fafb',
      border: '#000000',
      // Dark sections (WhyPhomiSection, footer) — same value as the
      // admin SideMenu background, for brand consistency (TZ §5).
      menu: '#333333',
    },
    status: {
      active: { bg: '#dcfce7', text: '#166534' },
      inactive: { bg: '#f3f4f6', text: '#374151' },
    },
  },
  // Public-site-only: large hero/section headings (TZ §5) — not needed in
  // the admin panel, which never shows full-bleed photo/video sections.
  typography: {
    hero: {
      fontSize: { base: '2.5rem', lg: '4rem', xl: '5rem' },
      fontWeight: '700',
      lineHeight: '1.05',
    },
  },
  radius: { sm: '4px', md: '8px', lg: '12px' },
  spacing: { unit: '4px' },
}

/** Public, binary status mapping for catalog cards — TZ §8.3. */
export const CATALOG_STATUS_MAP = {
  PRODUCT_STATUS_ACTIVE: 'active',
  PRODUCT_STATUS_DRAFT: 'inactive',
  PRODUCT_STATUS_INACTIVE: 'inactive',
  PRODUCT_STATUS_UNSPECIFIED: 'inactive',
}
