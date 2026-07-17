// Left side menu (TD §8.2): Dashboard, then every permission-visible
// entity grouped into logical sections, separated by a divider line (no
// section labels — just groups items visually). New entities show up
// under their section automatically once registered in
// config/entities/index.js with a `group` key — nothing to edit here
// when adding one to an existing group.
export const NAV_ITEMS = [{ key: 'dashboard', label: 'nav.dashboard', icon: 'i-lucide-layout-dashboard', to: '/' }]

// Order here controls the order sections render in the menu.
export const NAV_GROUPS = [{ key: 'sales' }, { key: 'catalog' }, { key: 'warehouse' }, { key: 'users' }]

// Curated menu subset per role, on top of the normal read-permission
// filter. Employee has read access to the full catalog (brands,
// categories, warehouses, product variants/SKUs, ...) because that data
// backs relation lookups elsewhere (e.g. a product card showing its
// brand name) — but only these five entities are meant to be entry
// points in their menu. A role with no entry here (admin, guest) gets
// every permission-visible entity, unrestricted.
export const ROLE_MENU_KEYS = {
  employee: ['clients', 'partners', 'sales', 'products', 'inventory'],
}
