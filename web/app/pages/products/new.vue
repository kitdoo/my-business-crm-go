<script setup>
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.products.create') }}</h1>
    <ProductGeneralForm :id="null" />
  </div>
</template>
