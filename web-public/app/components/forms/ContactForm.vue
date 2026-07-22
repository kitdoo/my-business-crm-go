<script setup>
import { contactFormSchema } from '~/utils/formSchemas.js'

const { t } = useI18n()
const route = useRoute()
const emit = defineEmits(['submitted'])

const state = reactive({
  name: '',
  email: '',
  phone: '',
  message: route.query.message?.toString() || '',
  website: '',
})
const submitting = ref(false)
const errorMessage = ref('')

async function onSubmit() {
  errorMessage.value = ''
  const parsed = contactFormSchema.safeParse(state)
  if (!parsed.success) {
    errorMessage.value = t('forms.validationError')
    return
  }
  submitting.value = true
  try {
    await $fetch('/api/forms/contact', { method: 'POST', body: parsed.data })
    emit('submitted')
  } catch {
    errorMessage.value = t('forms.submitError')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <form class="relative space-y-4" @submit.prevent="onSubmit">
    <div class="absolute left-0 top-0 h-0 w-0 overflow-hidden" aria-hidden="true">
      <label>Website<input v-model="state.website" type="text" tabindex="-1" autocomplete="off" /></label>
    </div>

    <UFormField :label="t('forms.contact.name')">
      <UInput v-model="state.name" class="w-full" required />
    </UFormField>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <UFormField :label="t('forms.contact.email')">
        <UInput v-model="state.email" type="email" class="w-full" required />
      </UFormField>
      <UFormField :label="t('forms.contact.phone')">
        <UInput v-model="state.phone" class="w-full" />
      </UFormField>
    </div>
    <UFormField :label="t('forms.contact.message')">
      <UTextarea v-model="state.message" class="w-full" :rows="5" required />
    </UFormField>

    <p v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</p>

    <UButton type="submit" variant="cta-outline" size="xl" block :loading="submitting">
      {{ t('forms.submit') }}
    </UButton>
  </form>
</template>
