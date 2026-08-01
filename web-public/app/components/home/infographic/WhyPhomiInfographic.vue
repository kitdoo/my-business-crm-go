<script setup>
// "Zašto Phomi?" infographic — see InfographicPanel.vue for the shared
// mechanics. Coordinates below are eyeballed against the original baked-in
// reference (public/images/why_phomi.jpg, 1030x1280) and reproduced here as
// a live HTML/SVG/i18n rebuild (TZ_Phomi §4) instead of a flat raster.
//
// Icon mapping note: animation_why_phomi/ only has 6 distinct icon designs
// for 7 concepts (flame1.png and glasses-insurance.png are the same
// crack/shatter glyph in two source colors — no dedicated flame icon
// exists). change.png's recycle-arrows stand in for "does not fade/change".
// Pairing below was corrected per user feedback: flame1.png (flame glyph)
// now covers "non-flammable", glasses-insurance.png (crack glyph) covers
// "unbreakable", and temperature.png (thermometer+sun) covers "frost
// resistant" instead.
const { t } = useI18n()

const ICONS = '/images/icons/animation_why_phomi/'

const items = computed(() => [
  {
    icon: ICONS + 'drop.png',
    label: t('home.whyPhomi.infographic.items.noWaterAbsorption'),
    badge: { x: 655, y: 130 },
    labelPos: { x: 705, y: 108, align: 'left' },
    line: [[655, 155], [655, 300]],
  },
  {
    icon: ICONS + 'bio.png',
    label: t('home.whyPhomi.infographic.items.notPlastic'),
    badge: { x: 230, y: 330 },
    labelPos: { x: 230, y: 385, align: 'center' },
    line: [[255, 330], [460, 330], [460, 380]],
  },
  {
    icon: ICONS + 'flame1.png',
    label: t('home.whyPhomi.infographic.items.nonFlammable'),
    badge: { x: 70, y: 600 },
    labelPos: { x: 70, y: 655, align: 'center' },
    line: [[95, 600], [390, 600]],
  },
  {
    icon: ICONS + 'glasses-insurance.png',
    label: t('home.whyPhomi.infographic.items.unbreakable'),
    badge: { x: 800, y: 610 },
    labelPos: { x: 800, y: 665, align: 'center' },
    line: [[800, 585], [700, 585], [700, 520]],
  },
  {
    icon: ICONS + 'temperature.png',
    label: t('home.whyPhomi.infographic.items.frostResistant'),
    badge: { x: 700, y: 855 },
    labelPos: { x: 740, y: 835, align: 'left' },
    line: [[700, 830], [600, 830], [600, 700]],
  },
  {
    icon: ICONS + 'lightweight.png',
    label: t('home.whyPhomi.infographic.items.lightThinFlexible'),
    badge: { x: 140, y: 1065 },
    labelPos: { x: 140, y: 1120, align: 'center' },
    line: [[140, 1040], [280, 1040], [280, 880]],
  },
  {
    icon: ICONS + 'change.png',
    label: t('home.whyPhomi.infographic.items.fadeResistant'),
    badge: { x: 570, y: 1025 },
    labelPos: { x: 570, y: 1080, align: 'center' },
    line: [[570, 1000], [570, 850], [520, 780]],
  },
])
</script>

<template>
  <InfographicPanel
    image="/images/animation/why_phomi.webp"
    :image-alt="t('home.whyPhomi.infographic.titleLine2')"
    :image-width="1920"
    :image-height="1908"
    :view-box-w="1030"
    :view-box-h="1280"
    :title-line1="t('home.whyPhomi.infographic.titleLine1')"
    :title-line2="t('home.whyPhomi.infographic.titleLine2')"
    :image-box="{ left: 6, top: 14, width: 90, height: 78 }"
    :items="items"
  />
</template>
