<script setup>
// Offer/receipt generation modal for a Sale (POST /invoices/{saleId}) —
// shared by the sales list (row action icon) and the sale detail page so
// clicking the icon in the list opens this in place instead of navigating
// to the detail page first (which used to fire the action *and* leave the
// list). Imperative `open(sale, documentType)` rather than props/v-model:
// the host doesn't otherwise need to track which sale is targeted.
const { t } = useI18n()
const clientApi = useEntityApi('clients')
const partnerApi = useEntityApi('partners')

const open = ref(false)
const sale = ref(null)
const documentType = ref('offer') // 'offer' | 'receipt' — see requestBody.documentType (server/api/invoices/[saleId].post.js)
const loading = ref(false)
const generating = ref(false)
const error = ref('')
const form = ref({
  taxId: '',
  registrationNumber: '',
  code: '',
  address: '',
  includeVat: false,
  sendEmail: false,
  email: '',
})
const title = computed(() =>
  t(documentType.value === 'receipt' ? 'entities.sales.generateReceipt' : 'entities.sales.generateInvoice'),
)

// Prefills from the sale's Client (or Partner, for a partner-only sale) —
// the list only loads the Sale itself, not the full buyer record, so this
// fetches it on open.
async function openModal(targetSale, docType) {
  sale.value = targetSale
  documentType.value = docType
  error.value = ''
  open.value = true
  loading.value = true
  try {
    const buyer = targetSale.clientId
      ? await clientApi.get(targetSale.clientId)
      : targetSale.partnerId
        ? await partnerApi.get(targetSale.partnerId)
        : null
    form.value = {
      taxId: buyer?.taxId || '',
      registrationNumber: buyer?.registrationNumber || '',
      code: buyer?.code || '',
      address: buyer?.address || '',
      includeVat: false,
      sendEmail: false,
      email: buyer?.email || '',
    }
  } catch (err) {
    error.value = err?.data?.error?.message || t('errors.generic')
  } finally {
    loading.value = false
  }
}

async function onGenerate() {
  generating.value = true
  error.value = ''
  try {
    const blob = await $fetch(`/api/invoices/${sale.value.id}`, {
      method: 'POST',
      body: {
        documentType: documentType.value,
        buyerTaxId: form.value.taxId,
        buyerRegistrationNumber: form.value.registrationNumber,
        buyerCode: form.value.code,
        buyerAddress: form.value.address,
        includeVat: form.value.includeVat,
        sendEmail: form.value.sendEmail,
        emailOverride: form.value.sendEmail ? form.value.email : undefined,
      },
      responseType: 'blob',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${documentType.value}-${sale.value.number}.pdf`
    link.click()
    URL.revokeObjectURL(url)
    open.value = false
  } catch (err) {
    error.value = err?.data?.error?.message || t('errors.generic')
  } finally {
    generating.value = false
  }
}

defineExpose({ open: openModal })
</script>

<template>
  <UModal v-model:open="open">
    <template #content>
      <div class="p-6 space-y-4">
        <h3 class="text-lg font-semibold">{{ title }}</h3>
        <div v-if="loading" class="py-4 text-center text-neutral-500">{{ t('common.loading') }}</div>
        <template v-else>
          <FormGrid>
            <UFormField :label="t('fields.taxId')">
              <UInput v-model="form.taxId" class="w-full" />
            </UFormField>
            <UFormField :label="t('fields.registrationNumber')">
              <UInput v-model="form.registrationNumber" class="w-full" />
            </UFormField>
            <UFormField :label="t('fields.code')">
              <UInput v-model="form.code" class="w-full" />
            </UFormField>
            <UFormField :label="t('fields.address')">
              <UInput v-model="form.address" class="w-full" />
            </UFormField>
          </FormGrid>
          <UCheckbox v-model="form.includeVat" :label="t('entities.sales.invoiceIncludeVat')" />
          <UCheckbox v-model="form.sendEmail" :label="t('entities.sales.invoiceSendEmail')" />
          <UFormField v-if="form.sendEmail" :label="t('fields.email')">
            <UInput v-model="form.email" type="email" class="w-full" />
          </UFormField>
        </template>
        <p v-if="error" class="text-sm text-error">{{ error }}</p>
        <div class="flex justify-end gap-2">
          <UButton color="neutral" variant="soft" @click="open = false">{{ t('common.cancel') }}</UButton>
          <UButton :disabled="loading" :loading="generating" @click="onGenerate">{{ title }}</UButton>
        </div>
      </div>
    </template>
  </UModal>
</template>
