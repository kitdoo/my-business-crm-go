<script setup>
definePageMeta({ layout: 'blank' })

const { t } = useI18n()
const { login } = useAuth()
const route = useRoute()
const router = useRouter()

const form = ref({ login: '', password: '' })
const error = ref('')
const loading = ref(false)

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    await login(form.value.login, form.value.password)
    router.push(route.query.redirect?.toString() || '/')
  } catch (err) {
    error.value = err?.data?.error?.message || t('login.error')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="relative min-h-screen flex">
    <div class="hidden md:block md:w-1/2 bg-[#333333]" />
    <div class="hidden md:block md:w-1/2 bg-brand-500" />

    <div class="absolute inset-0 flex items-center justify-center p-4">
      <div class="w-full max-w-sm rounded-lg border border-black bg-white p-6 shadow-lg">
        <h1 class="text-lg font-semibold mb-4 text-black">{{ t('login.title') }}</h1>
        <form class="space-y-4" @submit.prevent="onSubmit">
          <UFormField :label="t('login.login')">
            <UInput v-model="form.login" class="w-full" autocomplete="username" required />
          </UFormField>
          <UFormField :label="t('login.password')">
            <UInput v-model="form.password" type="password" class="w-full" autocomplete="current-password" required />
          </UFormField>
          <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
          <UButton type="submit" block :loading="loading">{{ t('login.submit') }}</UButton>
        </form>
      </div>
    </div>
  </div>
</template>
