<script setup>
// The only place LocalizedString editing logic lives (TD §5.2) — reused by
// every entity form with a LocalizedString field. v-model works with the
// plain { [locale]: string } shape; mapping to the proto { values: {...} }
// wrapper happens in useEntityForm, not here.
const props = defineProps({
  modelValue: { type: Object, required: true }, // { values: { [locale]: string } }
  locales: { type: Array, required: true },
  requiredLocale: { type: String, default: 'sr' },
  label: { type: String, default: '' },
  multiline: { type: Boolean, default: false },
  error: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])

// Locales to render: configured ones + any already present in the data,
// so an out-of-list locale from existing data is never silently dropped.
const allLocales = computed(() => {
  const existing = Object.keys(props.modelValue?.values || {})
  return Array.from(new Set([...props.locales, ...existing]))
})

const tabItems = computed(() => allLocales.value.map((locale) => ({ label: locale.toUpperCase(), value: locale })))
const activeTab = ref(props.requiredLocale)

function textFor(locale) {
  return props.modelValue?.values?.[locale] || ''
}

function updateLocale(locale, text) {
  const values = { ...(props.modelValue?.values || {}) }
  if (text) values[locale] = text
  else delete values[locale]
  emit('update:modelValue', { values })
}
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <UTabs v-model="activeTab" :items="tabItems" class="w-full">
      <template #content="{ item }">
        <UTextarea
          v-if="multiline"
          :model-value="textFor(item.value)"
          :placeholder="item.value === requiredLocale ? label : `${label} (${item.value})`"
          class="w-full mt-2"
          @update:model-value="(v) => updateLocale(item.value, v)"
        />
        <UInput
          v-else
          :model-value="textFor(item.value)"
          :placeholder="item.value === requiredLocale ? label : `${label} (${item.value})`"
          class="w-full mt-2"
          @update:model-value="(v) => updateLocale(item.value, v)"
        />
      </template>
    </UTabs>
  </UFormField>
</template>
