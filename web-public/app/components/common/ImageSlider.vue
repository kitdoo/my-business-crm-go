<script setup>
const props = defineProps({
  images: { type: Array, required: true },
  alt: { type: String, default: '' },
  caption: { type: String, default: '' },
  intervalMs: { type: Number, default: 5000 },
})

const active = ref(0)
let timer = null

function start() {
  stop()
  if (props.images.length < 2) return
  timer = setInterval(() => {
    active.value = (active.value + 1) % props.images.length
  }, props.intervalMs)
}

function stop() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

function goTo(i) {
  active.value = (i + props.images.length) % props.images.length
  start()
}

function next() {
  goTo(active.value + 1)
}

function prev() {
  goTo(active.value - 1)
}

onMounted(start)
onUnmounted(stop)
</script>

<template>
  <div class="group relative w-full h-full overflow-hidden rounded-sm bg-gray-50">
    <transition-group name="fade">
      <NuxtImg
        v-for="(src, i) in images"
        v-show="i === active"
        :key="src"
        :src="src"
        :alt="alt"
        loading="lazy"
        class="absolute inset-0 w-full h-full object-cover"
      />
    </transition-group>

    <template v-if="images.length > 1">
      <button
        type="button"
        aria-label="Previous"
        class="absolute left-2 top-1/2 -translate-y-1/2 w-8 h-8 flex items-center justify-center rounded-full bg-black/30 text-white opacity-0 group-hover:opacity-100 transition-opacity duration-200 hover:bg-black/50"
        @click="prev"
      >
        <UIcon name="i-lucide-chevron-left" class="w-5 h-5" />
      </button>
      <button
        type="button"
        aria-label="Next"
        class="absolute right-2 top-1/2 -translate-y-1/2 w-8 h-8 flex items-center justify-center rounded-full bg-black/30 text-white opacity-0 group-hover:opacity-100 transition-opacity duration-200 hover:bg-black/50"
        @click="next"
      >
        <UIcon name="i-lucide-chevron-right" class="w-5 h-5" />
      </button>
    </template>

    <div
      v-if="caption"
      class="absolute inset-x-0 bottom-0 px-3 py-2 bg-gradient-to-t from-black/60 to-transparent text-white text-sm opacity-0 group-hover:opacity-100 transition-opacity duration-200"
    >
      {{ caption }}
    </div>

    <div v-if="images.length > 1" class="absolute bottom-9 left-1/2 -translate-x-1/2 flex gap-1.5">
      <span
        v-for="(src, i) in images"
        :key="src"
        class="w-1.5 h-1.5 rounded-full transition-colors duration-300"
        :class="i === active ? 'bg-white' : 'bg-white/40'"
      />
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
