<script setup>
// Shared renderer for the two "peeled-apart" product infographics (Zašto
// Phomi? / Kako se ugrađuje Phomi?). Absolutely positioned callouts + an SVG
// line layer, all in percentage/viewBox coordinates so the whole panel
// (including badge size and label text) scales together with the container
// at every breakpoint — same line-pointer format on mobile, just smaller.
const props = defineProps({
  image: { type: String, required: true },
  imageAlt: { type: String, default: '' },
  // Native pixel dimensions of `image`, set as width/height attrs so the
  // browser can reserve layout space before the (lazy-loaded) file arrives.
  imageWidth: { type: Number, required: true },
  imageHeight: { type: Number, required: true },
  viewBoxW: { type: Number, required: true },
  viewBoxH: { type: Number, required: true },
  titleLine1: { type: String, required: true },
  titleLine2: { type: String, required: true },
  titleLine1Class: { type: String, default: 'text-brand-500' },
  titleLine2Class: { type: String, default: 'text-brand-500' },
  // Center artwork placement box, in % of the container — differs per
  // graphic since the two source PNGs have very different native crops.
  imageBox: {
    type: Object,
    default: () => ({ left: 6, top: 20, width: 88, height: 76 }),
  },
  items: { type: Array, required: true },
  /* item shape:
     { icon, label,
       badge: {x, y},               // viewBox units — badge center
       labelPos: {x, y, align},     // viewBox units — label anchor, align: left|center|right
       line: [[x,y], [x,y], ...] }  // viewBox units, starts at badge, ends on the image
  */
})

const { el, visible } = useRevealOnScroll()
const firstBadgeEl = ref(null)

// Clearance radius (viewBox units) a line must clear before it's allowed to
// bend — measured from the actual rendered CalloutBadge (not a copy of its
// CSS clamp() formula, so it can't drift out of sync) and converted via the
// panel's live px-per-unit ratio, so it tracks every breakpoint exactly.
const badgeRadiusUnits = ref(0)

function updateBadgeRadius() {
  if (!el.value || !firstBadgeEl.value) return
  const pxPerUnit = el.value.clientWidth / props.viewBoxW
  const badgeRadiusPx = firstBadgeEl.value.getBoundingClientRect().width / 2
  if (pxPerUnit > 0 && badgeRadiusPx > 0) {
    // +40% clearance so the line clears the icon's edge, not just its center.
    badgeRadiusUnits.value = (badgeRadiusPx / pxPerUnit) * 1.4
  }
}

let resizeObserver
onMounted(() => {
  updateBadgeRadius()
  resizeObserver = new ResizeObserver(updateBadgeRadius)
  if (el.value) resizeObserver.observe(el.value)
  window.addEventListener('resize', updateBadgeRadius)
})
onUnmounted(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('resize', updateBadgeRadius)
})

// Pushes the line's exit point out to the clearance circle's edge, straight
// out from the badge center along whichever side (top/bottom/left/right)
// the line already leaves from — never at an arbitrary angle. Axis-aligned
// callout lines (the only kind these infographics use) exit sharing either
// their x or y with the badge center; that shared coordinate is preserved,
// only the other one is pushed out.
//
// Critically, when the very next point sits on the same "shelf" as the
// original exit point (an immediate right-angle turn right next to the
// badge, e.g. glasses-insurance/lightweight), that point is pushed out by
// the same amount too. Moving only the first point in that case would leave
// the turn behind at the old, too-close height/position — turning a clean
// vertical-then-horizontal exit into a diagonal cutting back across the
// icon, which is exactly what still crossed the icon for those two items.
const clippedItems = computed(() =>
  props.items.map((item) => {
    const r = badgeRadiusUnits.value
    const [p0x, p0y] = item.line[0]
    const isVerticalExit = p0x === item.badge.x
    const isHorizontalExit = p0y === item.badge.y
    if (!r || (!isVerticalExit && !isHorizontalExit)) return item

    const axis = isVerticalExit ? 'y' : 'x'
    const oldVal = axis === 'y' ? p0y : p0x
    const anchor = axis === 'y' ? item.badge.y : item.badge.x
    const sign = Math.sign(oldVal - anchor) || 1
    const newVal = anchor + sign * r
    const delta = newVal - oldVal

    const line = item.line.map((pt) => [...pt])
    line[0][axis === 'y' ? 1 : 0] = newVal
    if (line[1] && line[1][axis === 'y' ? 1 : 0] === oldVal) {
      line[1][axis === 'y' ? 1 : 0] += delta
    }
    return { ...item, line }
  }),
)
</script>

<template>
  <section
    ref="el"
    class="relative w-full overflow-hidden infographic-panel"
    :style="{ '--vb-ratio': `${viewBoxW} / ${viewBoxH}` }"
  >
    <div class="absolute inset-0">
      <div
        class="absolute"
        :style="{
          left: imageBox.left + '%',
          top: imageBox.top + '%',
          width: imageBox.width + '%',
          height: imageBox.height + '%',
        }"
      >
        <img :src="image" :alt="imageAlt" :width="imageWidth" :height="imageHeight" class="w-full h-full object-contain" loading="lazy" />
      </div>

      <svg
        class="absolute inset-0 w-full h-full pointer-events-none"
        :viewBox.attr="`0 0 ${viewBoxW} ${viewBoxH}`"
        preserveAspectRatio="none"
      >
        <RevealLine
          v-for="(item, i) in clippedItems"
          :key="'line-' + i"
          :points="item.line"
          :visible="visible"
          :delay-ms="i * 90"
        />
      </svg>

      <div
        v-for="(item, i) in items"
        :key="'badge-' + i"
        class="absolute -translate-x-1/2 -translate-y-1/2 transition-all duration-500 ease-out"
        :class="visible ? 'opacity-100 scale-100' : 'opacity-0 scale-75'"
        :style="{
          left: (item.badge.x / viewBoxW) * 100 + '%',
          top: (item.badge.y / viewBoxH) * 100 + '%',
          transitionDelay: i * 90 + 200 + 'ms',
        }"
      >
        <CalloutBadge :ref="i === 0 ? (r) => (firstBadgeEl = r?.$el ?? r) : undefined" :icon="item.icon" />
      </div>

      <div
        v-for="(item, i) in items"
        :key="'label-' + i"
        class="absolute transition-all duration-500 ease-out"
        :class="[
          visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2',
          item.labelPos.align === 'left' ? 'text-left -translate-x-0' : item.labelPos.align === 'right' ? 'text-right -translate-x-full' : 'text-center -translate-x-1/2',
        ]"
        :style="{
          left: (item.labelPos.x / viewBoxW) * 100 + '%',
          top: (item.labelPos.y / viewBoxH) * 100 + '%',
          width: 'clamp(84px, 16vw, 190px)',
          transitionDelay: i * 90 + 320 + 'ms',
        }"
      >
        <span class="text-white font-bold uppercase leading-tight" style="font-size: clamp(11px, 1.6vw, 15px);">
          {{ item.label }}
        </span>
      </div>
    </div>

    <!-- Title -->
    <div class="absolute left-[5%] top-[5%]">
      <h2 class="font-extrabold uppercase leading-[1.05]" style="font-size: clamp(1.3rem, 4vw, 3rem);">
        <span class="block" :class="titleLine1Class">{{ titleLine1 }}</span>
        <span class="block" :class="titleLine2Class">{{ titleLine2 }}</span>
      </h2>
    </div>
  </section>
</template>

<style scoped>
/* Same annotated line-pointer format at every breakpoint (per user
   feedback) — the aspect ratio stays locked so badge/label/line positions
   (all % or viewBox based) scale down together instead of the container
   reflowing to a different layout on small screens. */
.infographic-panel {
  aspect-ratio: var(--vb-ratio);
}
</style>
