<script setup>
const route = useRoute()
</script>

<template>
  <CatalogPage :category-slug="route.params.category" />
</template>
