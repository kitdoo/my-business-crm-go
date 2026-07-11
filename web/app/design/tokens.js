/**
 * Single source of truth for design tokens (TD §7). Tailwind/@nuxt/ui theme
 * (app.config.js, assets/css/main.css) and components all read from here —
 * changing the brand color means editing this file only.
 */
export const tokens = {
  color: {
    // Turquoise — primary brand color (buttons, TopBar, active nav).
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
      // Menu / side navigation background (TopBar/SideMenu, TD §8).
      menu: '#333333',
    },
    status: {
      active: { bg: '#dcfce7', text: '#166534' },
      inactive: { bg: '#f3f4f6', text: '#374151' },
      warning: { bg: '#fef9c3', text: '#854d0e' },
      danger: { bg: '#fee2e2', text: '#991b1b' },
      info: { bg: '#dbeafe', text: '#1e40af' },
    },
  },
  radius: { sm: '4px', md: '8px', lg: '12px' },
  spacing: { unit: '4px' },
}

/**
 * Maps every project enum to one of the 5 status color families above.
 * <StatusBadge :map="STATUS_COLOR_MAP.brand" /> — see TD §7/§9.3.
 */
export const STATUS_COLOR_MAP = {
  brand: {
    BRAND_STATUS_ACTIVE: 'active',
    BRAND_STATUS_INACTIVE: 'inactive',
    BRAND_STATUS_UNSPECIFIED: 'inactive',
  },
  partner: {
    PARTNER_STATUS_ACTIVE: 'active',
    PARTNER_STATUS_INACTIVE: 'inactive',
    PARTNER_STATUS_UNSPECIFIED: 'inactive',
  },
}
