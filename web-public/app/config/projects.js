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

// Photo sets living in public/images/projects, named project_<n>_<i>.<ext>
const images = (n, count, ext = 'jpg') =>
  Array.from({ length: count }, (_, i) => `/images/projects/project_${n}_${i + 1}.${ext}`)

export const projects = [
  {
    id: 'p1',
    images: images(1, 6),
    titleKey: 'projects.items.p1',
    usedMaterials: placeholderMaterials('p1-used', 2),
    recommendedMaterials: placeholderMaterials('p1-rec', 1),
  },
  {
    id: 'p2',
    images: images(2, 6),
    titleKey: 'projects.items.p2',
    usedMaterials: placeholderMaterials('p2-used', 2),
    recommendedMaterials: placeholderMaterials('p2-rec', 1),
  },
  {
    id: 'p3',
    images: images(3, 4),
    titleKey: 'projects.items.p3',
    usedMaterials: placeholderMaterials('p3-used', 2),
    recommendedMaterials: placeholderMaterials('p3-rec', 1),
  },
  {
    id: 'p4',
    images: images(4, 8),
    titleKey: 'projects.items.p4',
    usedMaterials: placeholderMaterials('p4-used', 2),
    recommendedMaterials: placeholderMaterials('p4-rec', 1),
  },
  {
    id: 'p5',
    images: images(5, 5),
    titleKey: 'projects.items.p5',
    usedMaterials: placeholderMaterials('p5-used', 2),
    recommendedMaterials: placeholderMaterials('p5-rec', 1),
  },
  {
    id: 'p6',
    images: images(6, 4, 'jpeg'),
    titleKey: 'projects.items.p6',
    usedMaterials: placeholderMaterials('p6-used', 2),
    recommendedMaterials: placeholderMaterials('p6-rec', 1),
  },
  {
    id: 'p7',
    images: images(7, 7),
    titleKey: 'projects.items.p7',
    usedMaterials: placeholderMaterials('p7-used', 2),
    recommendedMaterials: placeholderMaterials('p7-rec', 1),
  },
]
