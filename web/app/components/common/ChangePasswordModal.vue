<script setup>
// Self-service password change (TD §4.1/§12.5) — any authenticated role,
// no permission gate (the current password itself proves the caller's
// right to act). Reachable from TopBar's profile menu.
const props = defineProps({
  open: { type: Boolean, default: false },
})
const emit = defineEmits(['update:open'])

const { t } = useI18n()
const toast = useToast()
const { handle } = useApiErrorHandler()

const currentPassword = ref('')
const newPassword = ref('')
const saving = ref(false)
const fieldErrors = ref({})

function reset() {
  currentPassword.value = ''
  newPassword.value = ''
  fieldErrors.value = {}
}

function onOpenChange(value) {
  emit('update:open', value)
  if (!value) reset()
}

async function onSubmit() {
  fieldErrors.value = {}
  saving.value = true
  try {
    await $fetch('/api/auth/change-password', {
      method: 'POST',
      body: { currentPassword: currentPassword.value, newPassword: newPassword.value },
    })
    toast.add({ title: t('topbar.changePasswordSuccess'), color: 'success' })
    onOpenChange(false)
  } catch (err) {
    handle(err, {
      onFieldError: (field, message) => {
        fieldErrors.value = { ...fieldErrors.value, [field]: message }
      },
    })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <UModal :open="open" @update:open="onOpenChange">
    <template #content>
      <div class="p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ t('topbar.changePassword') }}</h3>
        <form class="space-y-4" @submit.prevent="onSubmit">
          <UFormField :label="t('fields.currentPassword')" :error="fieldErrors.currentPassword" required>
            <UInput v-model="currentPassword" type="password" class="w-full" required autocomplete="current-password" />
          </UFormField>
          <UFormField :label="t('fields.newPassword')" :error="fieldErrors.newPassword" required>
            <UInput
              v-model="newPassword"
              type="password"
              class="w-full"
              required
              minlength="8"
              maxlength="128"
              autocomplete="new-password"
            />
          </UFormField>
          <div class="flex justify-end gap-2">
            <UButton color="neutral" variant="soft" @click="onOpenChange(false)">{{ t('common.cancel') }}</UButton>
            <UButton type="submit" :loading="saving">{{ t('common.save') }}</UButton>
          </div>
        </form>
      </div>
    </template>
  </UModal>
</template>
