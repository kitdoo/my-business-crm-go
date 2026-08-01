<script setup>
// Cancel-with-reason modal for a Sale — shared by the sales list (row
// action icon) and the sale detail page, same reasoning as
// SaleInvoiceModal.vue: opening in place instead of navigating away first.
const emit = defineEmits(['cancelled'])
const { t } = useI18n()
const saleApi = useSaleApi()
const { handle } = useApiErrorHandler()

const open = ref(false)
const sale = ref(null)
const reason = ref('')
const cancelling = ref(false)

function openModal(targetSale) {
  sale.value = targetSale
  reason.value = ''
  open.value = true
}

async function onConfirm() {
  cancelling.value = true
  try {
    const updated = await saleApi.cancel(sale.value.id, reason.value, sale.value.etag)
    open.value = false
    emit('cancelled', updated)
  } catch (err) {
    handle(err)
  } finally {
    cancelling.value = false
  }
}

defineExpose({ open: openModal })
</script>

<template>
  <UModal v-model:open="open">
    <template #content>
      <div class="p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ t('entities.sales.cancelConfirmTitle') }}</h3>
        <p class="text-sm text-neutral-600">{{ t('entities.sales.cancelConfirmBody') }}</p>
        <UFormField :label="t('entities.sales.cancelReason')">
          <UTextarea v-model="reason" class="w-full" />
        </UFormField>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="soft" @click="open = false">{{ t('common.back') }}</UButton>
          <UButton color="error" :loading="cancelling" @click="onConfirm">{{ t('entities.sales.cancel') }}</UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
