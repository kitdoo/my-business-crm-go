<script setup>
import { contactFormSchema } from '~/utils/formSchemas.js'

const { t } = useI18n()
const { items, isOpen, removeItem, setQty, clear, summaryText, totalAmount, totalCurrency } = useCart()

const showCheckout = ref(false)
const submitted = ref(false)
const submitting = ref(false)
const errorMessage = ref('')

const state = reactive({
  name: '',
  email: '',
  phone: '',
  message: '',
  website: '',
})

// The message textarea is seeded from the cart every time checkout opens,
// but stays editable — the visitor can add notes before sending (item 10).
watch(showCheckout, (val) => {
  if (val) state.message = summaryText.value
})

watch(isOpen, (val) => {
  if (!val) {
    showCheckout.value = false
    submitted.value = false
    errorMessage.value = ''
  }
})

async function onCheckoutSubmit() {
  errorMessage.value = ''
  const parsed = contactFormSchema.safeParse(state)
  if (!parsed.success) {
    errorMessage.value = t('forms.validationError')
    return
  }
  submitting.value = true
  try {
    await $fetch('/api/forms/contact', { method: 'POST', body: parsed.data })
    submitted.value = true
    clear()
  } catch {
    errorMessage.value = t('forms.submitError')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <USlideover v-model:open="isOpen" :title="showCheckout ? t('cart.checkoutTitle') : t('cart.title')">
    <template #body>
      <div v-if="submitted" class="text-center py-10">
        <UIcon name="i-lucide-circle-check" class="w-14 h-14 text-brand-500 mx-auto mb-6" />
        <p class="text-lg">{{ t('cart.thanks') }}</p>
      </div>

      <form v-else-if="showCheckout" class="relative space-y-4" @submit.prevent="onCheckoutSubmit">
        <div class="absolute -left-[9999px]" aria-hidden="true">
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
        <UFormField :label="t('cart.messageLabel')">
          <UTextarea v-model="state.message" class="w-full" :rows="6" required />
        </UFormField>

        <p v-if="errorMessage" class="text-sm text-red-600">{{ errorMessage }}</p>

        <div class="flex gap-3">
          <UButton variant="cta-outline" class="px-6" @click="showCheckout = false">{{ t('cart.back') }}</UButton>
          <UButton type="submit" variant="cta-outline" class="flex-1" :loading="submitting">{{ t('cart.send') }}</UButton>
        </div>
      </form>

      <div v-else>
        <p v-if="!items.length" class="text-black/50 py-10 text-center">{{ t('cart.empty') }}</p>
        <ul v-else class="space-y-4">
          <li v-for="item in items" :key="item.sku" class="flex items-center gap-3 border-b border-black/10 pb-4">
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ item.name }}</p>
              <p v-if="item.options" class="text-xs text-black/50 truncate">{{ item.options }}</p>
              <p class="text-xs text-black/40">{{ item.sku }}</p>
              <MoneyLabel
                v-if="item.price"
                :amount="item.price.amount * item.qty"
                :currency="item.price.currency"
                class="text-xs text-brand-700 font-medium"
              />
            </div>
            <UInput
              :model-value="item.qty"
              type="number"
              min="1"
              class="w-16"
              @update:model-value="(v) => setQty(item.sku, Number(v))"
            />
            <button :aria-label="t('cart.remove')" class="text-black/40 hover:text-red-600" @click="removeItem(item.sku)">
              <UIcon name="i-lucide-x" class="w-5 h-5" />
            </button>
          </li>
        </ul>

        <div v-if="items.length && totalCurrency" class="flex items-center justify-between mt-6 pt-4 border-t border-black/10">
          <span class="font-medium uppercase text-sm tracking-wide">{{ t('cart.total') }}</span>
          <MoneyLabel :amount="totalAmount" :currency="totalCurrency" class="text-lg font-semibold text-brand-700" />
        </div>

        <UButton
          v-if="items.length"
          variant="cta-outline"
          size="xl"
          class="w-full mt-6 py-4 text-base"
          @click="showCheckout = true"
        >
          {{ t('cart.checkout') }}
        </UButton>
      </div>
    </template>
  </USlideover>
</template>
