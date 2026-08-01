<script setup>
// "Kako se ugrađuje Phomi?" infographic — see InfographicPanel.vue for the
// shared mechanics. Coordinates eyeballed against the original baked-in
// reference (public/images/how_to_install.jpg, 1280x1280), TZ_Phomi §5, then
// corrected per user feedback to point at the actual layer in the exploded
// wall diagram (counting layers left→right: 1 brick, 2 base mesh, 3 adhesive
// coat, 4 mineral wool insulation, 5 base coat, 6 reinforcing mesh 160g/m²,
// 7 base coat, 8 outdoor tile adhesive (flat, pale coat), 9 Phomi panel +
// grout — layer 9 is drawn as two abutting stone-textured tiles with a thin
// grout seam between them, both still the same final decorative layer).
// grid.png (mesh glyph) and solar-panel.png were swapped between
// reinforcementLayer/phomiLayer so the mesh icon matches the mesh layer.
const { t } = useI18n()

const ICONS = '/images/icons/animation_install/'

const items = computed(() => [
  {
    icon: ICONS + 'wall.png',
    label: t('home.install.infographic.items.applyingAdhesive'),
    badge: { x: 256, y: 340 },
    labelPos: { x: 256, y: 250, align: 'center' },
    line: [[256, 365], [256, 550], [351, 550], [351, 700]],
  },
  {
    icon: ICONS + 'insulation1.png',
    label: t('home.install.infographic.items.installingInsulation'),
    badge: { x: 678, y: 290 },
    labelPos: { x: 678, y: 200, align: 'center' },
    line: [[678, 315], [678, 550], [470, 550], [470, 750]],
  },
  {
    icon: ICONS + 'maintenance.png',
    label: t('home.install.infographic.items.fixingInsulation'),
    badge: { x: 900, y: 340 },
    labelPos: { x: 940, y: 250, align: 'left' },
    line: [[900, 365], [900, 500], [744, 500], [744, 601]],
  },
  {
    icon: ICONS + 'solar-panel.png',
    label: t('home.install.infographic.items.phomiLayer'),
    badge: { x: 1180, y: 505 },
    labelPos: { x: 1180, y: 415, align: 'center' },
    line: [[1180, 530], [1170, 700]],
  },
  {
    icon: ICONS + 'grid.png',
    label: t('home.install.infographic.items.reinforcementLayer'),
    badge: { x: 450, y: 1150 },
    labelPos: { x: 450, y: 1215, align: 'center' },
    line: [[450, 1125], [450, 858], [693, 858]],
  },
  {
    icon: ICONS + 'trowel.png',
    label: t('home.install.infographic.items.outdoorTileAdhesive'),
    badge: { x: 700, y: 1150 },
    labelPos: { x: 700, y: 1215, align: 'center' },
    line: [[700, 1125], [700, 974], [873, 974]],
  },
  {
    icon: ICONS + 'home.png',
    label: t('home.install.infographic.items.grouting'),
    badge: { x: 1150, y: 1150 },
    labelPos: { x: 1150, y: 1215, align: 'center' },
    line: [[1150, 1125], [1150, 1000], [1205, 1000], [1205, 858]],
  },
])
</script>

<template>
  <InfographicPanel
    image="/images/animation/install.webp"
    :image-alt="t('home.install.infographic.titleLine2')"
    :image-width="1813"
    :image-height="1109"
    :view-box-w="1280"
    :view-box-h="1280"
    :title-line1="t('home.install.infographic.titleLine1')"
    :title-line2="t('home.install.infographic.titleLine2')"
    title-line1-class="text-white"
    :image-box="{ left: 2, top: 32, width: 96, height: 60 }"
    :items="items"
  />
</template>
