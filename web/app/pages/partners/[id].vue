<script setup>
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.partners.edit') }}</h1>
    <EntityForm entity="partners" :id="route.params.id" />
  </div>
</template>
