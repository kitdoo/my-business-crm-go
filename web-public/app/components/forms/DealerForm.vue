<script setup>
import { dealerFormSchema } from '~/utils/formSchemas.js'

const { t } = useI18n()
const emit = defineEmits(['submitted'])

const state = reactive({
  companyName: '',
  contactName: '',
  phone: '',
  email: '',
  city: '',
  message: '',
  website: '', // honeypot — kept hidden via CSS below
})
const submitting = ref(false)
const errorMessage = ref('')

async function onSubmit() {
  errorMessage.value = ''
  const parsed = dealerFormSchema.safeParse(state)
  if (!parsed.success) {
    errorMessage.value = t('forms.validationError')
    return
  }
  submitting.value = true
  try {
    await $fetch('/api/forms/dealer', { method: 'POST', body: parsed.data })
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
    <div class="absolute -left-[9999px]" aria-hidden="true">
      <label>Website<input v-model="state.website" type="text" tabindex="-1" autocomplete="off" /></label>
    </div>

    <UFormField :label="t('forms.dealer.companyName')">
      <UInput v-model="state.companyName" class="w-full" required />
    </UFormField>
    <UFormField :label="t('forms.dealer.contactName')">
      <UInput v-model="state.contactName" class="w-full" required />
    </UFormField>
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <UFormField :label="t('forms.dealer.phone')">
        <UInput v-model="state.phone" class="w-full" />
      </UFormField>
      <UFormField :label="t('forms.dealer.email')">
        <UInput v-model="state.email" type="email" class="w-full" required />
      </UFormField>
    </div>
    <UFormField :label="t('forms.dealer.city')">
      <UInput v-model="state.city" class="w-full" required />
    </UFormField>
    <UFormField :label="t('forms.dealer.message')">
      <UTextarea v-model="state.message" class="w-full" :rows="4" />
    </UFormField>

    <p v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</p>

    <UButton type="submit" variant="cta-outline" size="xl" block :loading="submitting">
      {{ t('forms.submit') }}
    </UButton>
  </form>
</template>
