<script setup>
// Read-only "ids -> comma-joined names" for a repeated relation field
// (e.g. Product.categoryIds) — same resolution as RelationLabel (TD
// §9.3/§9.5), just one per id.
const props = defineProps({
  value: { type: Array, default: () => [] },
  relation: { type: String, required: true },
})
</script>

<template>
  <span v-if="!value?.length" />
  <span v-else>
    <template v-for="(id, index) in value" :key="id">
      <RelationLabel :value="id" :relation="relation" />{{ index < value.length - 1 ? ', ' : '' }}
    </template>
  </span>
</template>
