// "Završeni projekti" data — not a CRM entity (TZ §4.3/§9), site content
// only. Only real project photos are listed here — no placeholder/SVG
// projects until real photos are provided for them.
//
// materials are the real material photos used on each project, shown via
// MaterialsCollage.vue with a baked name/color/size label per tile — no
// catalog SKU link yet, just documentation of what was used.

// Photo sets live in public/images/projects/<n>/, named project_<n>_<i>.<ext>
const images = (n, count, ext = 'jpg') =>
  Array.from({ length: count }, (_, i) => `/images/projects/${n}/project_${n}_${i + 1}.${ext}`)

// Material photos live alongside the gallery photos, named
// project_<n>_material_<i>.<ext> — extension/case matches the actual file
// on disk (the production filesystem is case-sensitive).
const materialImage = (n, i, ext) => `/images/projects/${n}/project_${n}_material_${i}.${ext}`

export const projects = [
  {
    id: 'p1',
    images: images(1, 6),
    titleKey: 'projects.items.p1',
    materials: [
      { image: materialImage(1, 1, 'JPG'), name: 'POLY WOOD', color: 'PEACH', size: '300×2700' },
      { image: materialImage(1, 2, 'jpg'), name: 'MOUNT CELESTIAL', color: 'H06', size: '1200×600' },
      { image: materialImage(1, 3, 'JPG'), name: 'POLISH CONCRETE STONE', color: 'GREYISH DESERT', size: '1200×600' },
    ],
  },
  {
    id: 'p2',
    images: images(2, 6),
    titleKey: 'projects.items.p2',
    materials: [
      { image: materialImage(2, 1, 'JPG'), name: 'ROUGH SURFACE', color: 'PORTORO', size: '1200×600' },
    ],
  },
  {
    id: 'p3',
    images: images(3, 4),
    titleKey: 'projects.items.p3',
    materials: [
      { image: materialImage(3, 1, 'jpg'), name: 'MARBLE', color: 'CASTOL GREY', size: '1200×600' },
    ],
  },
  {
    id: 'p4',
    images: images(4, 8),
    titleKey: 'projects.items.p4',
    materials: [
      { image: materialImage(4, 1, 'JPG'), name: 'ROUGH SURFACE', color: 'PORTORO', size: '1200×600' },
      { image: materialImage(4, 2, 'jpg'), name: 'MOUNT CELESTIAL', color: 'H06', size: '1200×600' },
      { image: materialImage(4, 3, 'JPG'), name: 'POLISH CONCRETE STONE', color: 'GREYISH DESERT', size: '1200×600' },
    ],
  },
  {
    id: 'p5',
    images: images(5, 5),
    titleKey: 'projects.items.p5',
    materials: [
      { image: materialImage(5, 1, 'JPG'), name: 'ROME TRAVERTINE', color: 'CLOUD GREY', size: '1200×600' },
      { image: materialImage(5, 2, 'JPG'), name: 'ROUGH SURFACE', color: 'PORTORO', size: '1200×600' },
    ],
  },
  {
    id: 'p6',
    images: images(6, 4, 'jpeg'),
    titleKey: 'projects.items.p6',
    materials: [
      { image: materialImage(6, 1, 'JPG'), name: 'ROME TRAVERTINE', color: 'CLOUD YELLOW', size: '' },
      { image: materialImage(6, 2, 'jpg'), name: 'CHISELED STONE', color: 'CLOUD YELLOW', size: '' },
    ],
  },
  {
    id: 'p7',
    images: images(7, 7),
    titleKey: 'projects.items.p7',
    materials: [
      { image: materialImage(7, 1, 'JPG'), name: 'ORIGINAL WOOD', color: 'DARK BROWN', size: '1200×600' },
      { image: materialImage(7, 2, 'JPG'), name: 'POLY WOOD', color: 'PEACH', size: '300×2700' },
      { image: materialImage(7, 3, 'JPG'), name: 'MOUNT CELESTIAL', color: 'MEDIUM GREY', size: '1200×600' },
      { image: materialImage(7, 4, 'jpg'), name: 'MOUNT CELESTIAL', color: 'H06', size: '1200×600' },
    ],
  },
]
