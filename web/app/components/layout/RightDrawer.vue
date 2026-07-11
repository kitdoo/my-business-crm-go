<script setup>
// Flat side menu (TD §8.2): Dashboard + every permission-visible entity,
// one level, no intermediate section click. Permanent panel on >= lg,
// overlay slideover below that (TD §6).
import { NAV_ITEMS } from '~/config/navigation.js'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { can } = usePermission()
const { rightDrawerOpen } = useLayoutState()

const navItems = computed(() => [
  ...NAV_ITEMS,
  ...listEntityConfigs()
    .filter((cfg) => can(cfg.permissions.read))
    .map((cfg) => ({ key: cfg.key, label: cfg.label, icon: cfg.icon, to: cfg.route })),
])

function isActive(item) {
  return item.to === '/' ? route.path === '/' : route.path.startsWith(item.to)
}

function onSelect(item) {
  router.push(item.to)
  rightDrawerOpen.value = false
}
</script>

<template>
  <nav class="w-56 shrink-0 h-full bg-[#333333] p-2 space-y-1">
    <button
      v-for="item in navItems"
      :key="item.key"
      class="w-full flex items-center gap-2 rounded-md px-3 py-2 text-sm text-left text-neutral-300 hover:bg-white/10 hover:text-white"
      :class="{ 'bg-brand-500 text-white font-medium hover:bg-brand-500': isActive(item) }"
      @click="onSelect(item)"
    >
      <UIcon :name="item.icon" class="size-4" />
      {{ t(item.label) }}
    </button>
  </nav>
</template>
