<script setup>
// Role-aware workflow guide (TD §8.2's Help entry). A page, not a modal —
// reachable from SideMenu, works the same whether opened via the menu or
// linked to directly. Reads help.roles.<role>.{intro,steps} from i18n.
const { t, tm, rt } = useI18n()
const { user } = useAuth()

function normalizeRole(role) {
  if (!role) return null
  return role.replace(/^USER_ROLE_/, '').toLowerCase()
}

const role = computed(() => normalizeRole(user.value?.role))
// tm() returns the raw message tree — leaf strings are compiled message
// functions, not plain strings, so each one needs rt() to render (TD n/a,
// this is a vue-i18n API contract: https://vue-i18n.intlify.dev/api/composition.html#tm).
const steps = computed(() => {
  if (!role.value) return []
  const raw = tm(`help.roles.${role.value}.steps`) ?? []
  return raw.map((step) => ({ icon: rt(step.icon), title: rt(step.title), description: rt(step.description) }))
})
const intro = computed(() => (role.value ? t(`help.roles.${role.value}.intro`) : t('help.noRole')))
</script>

<template>
  <div class="space-y-4 max-w-2xl">
    <h1 class="text-xl font-semibold">{{ t('help.title') }}</h1>
    <p class="text-sm text-neutral-500">{{ intro }}</p>

    <ol v-if="steps.length" class="space-y-3">
      <li v-for="(step, index) in steps" :key="index" class="flex gap-3">
        <div class="shrink-0 size-8 rounded-full bg-brand-500/10 text-brand-500 flex items-center justify-center">
          <UIcon :name="step.icon" class="size-4" />
        </div>
        <div>
          <div class="text-sm font-medium">{{ step.title }}</div>
          <div class="text-sm text-neutral-500">{{ step.description }}</div>
        </div>
      </li>
    </ol>
  </div>
</template>
