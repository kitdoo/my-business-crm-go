<script setup>
// Side menu (TD §8.2): Dashboard, then every permission-visible entity
// grouped into logical sections (sales, catalog, warehouse, users),
// separated by a plain divider line (no section labels), plus a Help
// entry linking to the /help page (role-aware workflow guide). Sits on
// the left; permanent panel on >= lg, overlay slideover below that (TD §6).
import { NAV_ITEMS, NAV_GROUPS } from '~/config/navigation.js'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { can } = usePermission()
const { menuOpen } = useLayoutState()

const visibleEntities = computed(() =>
  listEntityConfigs()
    .filter((cfg) => can(cfg.permissions.read))
    .map((cfg) => ({ key: cfg.key, label: cfg.label, icon: cfg.icon, to: cfg.route, group: cfg.group })),
)

const groupedSections = computed(() =>
  NAV_GROUPS.map((group) => ({
    key: group.key,
    items: visibleEntities.value.filter((item) => item.group === group.key),
  })).filter((group) => group.items.length > 0),
)

function isActive(item) {
  if (item.to === '/') return route.path === '/'
  return route.path === item.to || route.path.startsWith(`${item.to}/`)
}

function onSelect(item) {
  router.push(item.to)
  menuOpen.value = false
}
</script>

<template>
  <nav class="w-56 shrink-0 h-full bg-[#333333] p-2 flex flex-col">
    <div class="space-y-1">
      <button
        v-for="item in NAV_ITEMS"
        :key="item.key"
        class="w-full flex items-center gap-2 rounded-md px-3 py-2 text-sm text-left text-neutral-300 hover:bg-white/10 hover:text-white"
        :class="{ 'bg-brand-500 text-white font-medium hover:bg-brand-500': isActive(item) }"
        @click="onSelect(item)"
      >
        <UIcon :name="item.icon" class="size-4" />
        {{ t(item.label) }}
      </button>
    </div>

    <div v-for="section in groupedSections" :key="section.key" class="mt-2 pt-2 border-t border-white/10 space-y-1">
      <button
        v-for="item in section.items"
        :key="item.key"
        class="w-full flex items-center gap-2 rounded-md px-3 py-2 text-sm text-left text-neutral-300 hover:bg-white/10 hover:text-white"
        :class="{ 'bg-brand-500 text-white font-medium hover:bg-brand-500': isActive(item) }"
        @click="onSelect(item)"
      >
        <UIcon :name="item.icon" class="size-4" />
        {{ t(item.label) }}
      </button>
    </div>

    <div class="mt-auto pt-2 border-t border-white/10">
      <button
        class="w-full flex items-center gap-2 rounded-md px-3 py-2 text-sm text-left text-neutral-300 hover:bg-white/10 hover:text-white"
        :class="{ 'bg-brand-500 text-white font-medium hover:bg-brand-500': isActive({ to: '/help' }) }"
        @click="onSelect({ to: '/help' })"
      >
        <UIcon name="i-lucide-circle-help" class="size-4" />
        {{ t('nav.help') }}
      </button>
    </div>
  </nav>
</template>
