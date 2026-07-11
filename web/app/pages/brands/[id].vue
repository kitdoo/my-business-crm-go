<script setup>
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.brands.edit') }}</h1>
    <EntityForm entity="brands" :id="route.params.id" />
  </div>
</template>
