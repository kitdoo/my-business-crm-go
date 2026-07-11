<script setup>
// Single confirmation modal for every destructive/irreversible action
// (TD §9.3) — always states the consequence, never a bare "Are you sure?".
const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, required: true },
  description: { type: String, required: true },
  confirmLabel: { type: String, default: '' },
  danger: { type: Boolean, default: true },
})
const emit = defineEmits(['update:open', 'confirm', 'cancel'])
const { t } = useI18n()

function onConfirm() {
  emit('confirm')
  emit('update:open', false)
}
function onCancel() {
  emit('cancel')
  emit('update:open', false)
}
</script>

<template>
  <UModal :open="open" @update:open="(v) => emit('update:open', v)">
    <template #content>
      <div class="p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ title }}</h3>
        <p class="text-sm text-neutral-600">{{ description }}</p>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="soft" @click="onCancel">{{ t('common.cancel') }}</UButton>
          <UButton :color="danger ? 'error' : 'primary'" @click="onConfirm">
            {{ confirmLabel || t('common.confirm') }}
          </UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
