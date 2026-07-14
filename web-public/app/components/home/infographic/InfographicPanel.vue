<script setup>
// Shared renderer for the two "peeled-apart" product infographics (Zašto
// Phomi? / Kako se ugrađuje Phomi?). Absolutely positioned callouts + an SVG
// line layer, all in percentage/viewBox coordinates so the whole panel
// (including badge size and label text) scales together with the container
// at every breakpoint — same line-pointer format on mobile, just smaller.
defineProps({
  image: { type: String, required: true },
  imageAlt: { type: String, default: '' },
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
        <img :src="image" :alt="imageAlt" class="w-full h-full object-contain" loading="lazy" />
      </div>

      <svg
        class="absolute inset-0 w-full h-full pointer-events-none"
        :viewBox.attr="`0 0 ${viewBoxW} ${viewBoxH}`"
        preserveAspectRatio="none"
      >
        <RevealLine
          v-for="(item, i) in items"
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
        <CalloutBadge :icon="item.icon" />
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
