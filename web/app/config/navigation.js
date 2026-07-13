// Left side menu (TD §8.2): Dashboard, then every permission-visible
// entity grouped into logical sections, separated by a divider line (no
// section labels — just groups items visually). New entities show up
// under their section automatically once registered in
// config/entities/index.js with a `group` key — nothing to edit here
// when adding one to an existing group.
export const NAV_ITEMS = [{ key: 'dashboard', label: 'nav.dashboard', icon: 'i-lucide-layout-dashboard', to: '/' }]

// Order here controls the order sections render in the menu.
export const NAV_GROUPS = [{ key: 'sales' }, { key: 'catalog' }, { key: 'warehouse' }, { key: 'users' }]
