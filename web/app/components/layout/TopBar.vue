<script setup>
// Turquoise header (TD §8.1): logo -> dashboard, language switcher, profile
// menu (change password modal / logout). Hamburger on < lg opens SideMenu.
const { t, locale, locales, setLocale } = useI18n()
const { user, logout } = useAuth()
const { menuOpen } = useLayoutState()
const router = useRouter()
const runtimeConfig = useRuntimeConfig()

const changePasswordOpen = ref(false)

const profileItems = computed(() => [
  [{ label: t('topbar.changePassword'), icon: 'i-lucide-key-round', onSelect: () => (changePasswordOpen.value = true) }],
  [{ label: t('topbar.logout'), icon: 'i-lucide-log-out', onSelect: onLogout }],
])

const languageItems = computed(() =>
  locales.value.map((l) => [{ label: l.code.toUpperCase(), onSelect: () => setLocale(l.code) }]),
)

async function onLogout() {
  await logout()
  router.push('/login')
}
</script>

<template>
  <header class="h-14 flex items-center justify-between px-4 text-white bg-brand-500">
    <div class="flex items-center gap-3">
      <UButton
        icon="i-lucide-menu"
        color="neutral"
        variant="ghost"
        class="lg:hidden text-white"
        @click="menuOpen = !menuOpen"
      />
      <NuxtLink to="/" class="font-semibold text-white">{{ runtimeConfig.public.appName }}</NuxtLink>
    </div>

    <div class="flex items-center gap-2">
      <UDropdownMenu :items="languageItems">
        <UButton color="neutral" variant="ghost" class="text-white" icon="i-lucide-globe">
          {{ locale.toUpperCase() }}
        </UButton>
      </UDropdownMenu>

      <UDropdownMenu v-if="user" :items="profileItems">
        <UButton color="neutral" variant="ghost" class="text-white" icon="i-lucide-user">
          {{ user.name?.values?.sr || user.email }}
        </UButton>
      </UDropdownMenu>
    </div>

    <ChangePasswordModal v-model:open="changePasswordOpen" />
  </header>
</template>
