<script setup>
// Just the "arbitrary from/to range" control (TD §9.3) — a button that
// opens a small popover with two date inputs. This is the piece
// PeriodFilter's "custom" button and every entity list's periodFilter
// filter both need; PeriodFilter's today/7d/month presets are dashboard-
// only and don't belong in a table filter bar, so they're not part of
// this component.
const props = defineProps({
  modelValue: { type: Object, default: () => ({}) }, // { from, to } unix seconds
  active: { type: Boolean, default: false }, // visual "this range is applied" state
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

const open = ref(false)
const from = ref('')
const to = ref('')

function apply() {
  if (!from.value || !to.value) return
  const fromSeconds = Math.floor(new Date(`${from.value}T00:00:00`).getTime() / 1000)
  const toSeconds = Math.floor(new Date(`${to.value}T23:59:59`).getTime() / 1000)
  open.value = false
  emit('update:modelValue', { from: fromSeconds, to: toSeconds })
}
</script>

<template>
  <UPopover v-model:open="open">
    <UButton :variant="active ? 'solid' : 'soft'" size="sm" icon="i-lucide-calendar-range">
      {{ t('periodFilter.custom') }}
    </UButton>
    <template #content>
      <div class="p-4 space-y-3 w-64">
        <UFormField :label="t('periodFilter.from')">
          <UInput v-model="from" type="date" class="w-full" />
        </UFormField>
        <UFormField :label="t('periodFilter.to')">
          <UInput v-model="to" type="date" class="w-full" />
        </UFormField>
        <UButton block :disabled="!from || !to" @click="apply">{{ t('common.confirm') }}</UButton>
      </div>
    </template>
  </UPopover>
</template>
