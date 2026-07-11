<script setup>
// TD §12.1 describes a tabbed detail page (General/Price/Inventory/
// Movements) — the Price/Inventory/Movements tabs land once those
// entities exist (tasks #9-#11); for now this is the same single-form
// page every other entity gets, with images already part of the form.
const route = useRoute()
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <h1 class="text-xl font-semibold">{{ t('entities.products.edit') }}</h1>
    <EntityForm entity="products" :id="route.params.id" />
  </div>
</template>
