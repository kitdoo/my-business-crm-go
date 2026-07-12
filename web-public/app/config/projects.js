// "Završeni projekti" data — not a CRM entity (TZ §4.3/§9), site content
// only. Only real project photos are listed here — no placeholder/SVG
// projects until real photos are provided for them.
//
// usedMaterials/recommendedMaterials are placeholder material tiles (no
// real catalog SKU tied to a project yet) — they link to the catalog so
// visitors can browse, but aren't wired to a specific product page. Each
// project shows exactly 3 tiles, top to bottom: 2 used + 1 recommended/
// similar.
const PLACEHOLDER_MATERIAL = '/images/product-placeholder.svg'
const placeholderMaterials = (prefix, count) =>
  Array.from({ length: count }, (_, i) => ({ id: `${prefix}-${i + 1}`, image: PLACEHOLDER_MATERIAL }))

export const projects = [
  {
    id: 'p1',
    image: '/images/project_1.jpg',
    titleKey: 'projects.items.p1',
    usedMaterials: placeholderMaterials('p1-used', 2),
    recommendedMaterials: placeholderMaterials('p1-rec', 1),
  },
  {
    id: 'p2',
    image: '/images/project_2.jpg',
    titleKey: 'projects.items.p2',
    usedMaterials: placeholderMaterials('p2-used', 2),
    recommendedMaterials: placeholderMaterials('p2-rec', 1),
  },
]
