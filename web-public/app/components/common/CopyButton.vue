<script setup>
import { useClipboard } from '@vueuse/core'

const props = defineProps({
  value: { type: String, required: true },
})

const { t } = useI18n()
const { copy, copied } = useClipboard({ source: props.value, copiedDuring: 1500 })
</script>

<template>
  <button
    type="button"
    class="inline-flex items-center justify-center w-6 h-6 rounded-sm text-black/40 hover:text-brand-600 hover:bg-black/5 transition-colors"
    :aria-label="t('contact.copy')"
    :title="t('contact.copy')"
    @click="copy(value)"
  >
    <UIcon :name="copied ? 'i-lucide-check' : 'i-lucide-copy'" class="w-3.5 h-3.5" />
  </button>
</template>
