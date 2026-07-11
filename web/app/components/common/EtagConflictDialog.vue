<script setup>
// Shown whenever a mutation comes back `Aborted` (etag mismatch) — someone
// else changed the record while it was open (TD §4.4/§9.3).
defineProps({
  open: { type: Boolean, default: false },
})
const emit = defineEmits(['update:open', 'reload'])
const { t } = useI18n()

function onReload() {
  emit('reload')
  emit('update:open', false)
}
</script>

<template>
  <UModal :open="open" @update:open="(v) => emit('update:open', v)">
    <template #content>
      <div class="p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ t('errors.etagConflictTitle') }}</h3>
        <p class="text-sm text-neutral-600">{{ t('errors.etagConflictBody') }}</p>
        <div class="flex justify-end gap-2">
          <UButton color="primary" @click="onReload">{{ t('errors.etagConflictReload') }}</UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
