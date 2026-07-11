<script setup>
// Slot-based alternative to `v-if="can(...)"` for larger blocks (TD §9.3).
const props = defineProps({
  permission: { type: String, required: true },
})
const { can } = usePermission()
const allowed = computed(() => can(props.permission))
</script>

<template>
  <slot v-if="allowed" />
</template>
