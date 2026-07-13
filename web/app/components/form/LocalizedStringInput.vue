<script setup>
// The only place LocalizedString editing logic lives (TD §5.2) — reused by
// every entity form with a LocalizedString field. v-model works with the
// plain { [locale]: string } shape; mapping to the proto { values: {...} }
// wrapper happens in useEntityForm, not here.
//
// One row per locale, starting with just the required one — a tab per
// configured locale was confusing when most fields only ever need the
// default language. "+ add language" reveals another row on demand; a
// locale that already carries data (loaded from an existing record) is
// always shown, even if it's not on the picker list yet, so nothing is
// silently dropped.
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
const { t } = useI18n()

const shownLocales = ref(
  Array.from(new Set([props.requiredLocale, ...Object.keys(props.modelValue?.values || {})])),
)

const availableLocales = computed(() => props.locales.filter((l) => !shownLocales.value.includes(l)))
const addLocaleItems = computed(() => [
  availableLocales.value.map((locale) => ({ label: locale.toUpperCase(), onSelect: () => addLocale(locale) })),
])

function textFor(locale) {
  return props.modelValue?.values?.[locale] || ''
}

function updateLocale(locale, text) {
  const values = { ...(props.modelValue?.values || {}) }
  if (text) values[locale] = text
  else delete values[locale]
  emit('update:modelValue', { values })
}

function addLocale(locale) {
  if (!locale || shownLocales.value.includes(locale)) return
  shownLocales.value = [...shownLocales.value, locale]
}

function removeLocale(locale) {
  if (locale === props.requiredLocale) return
  shownLocales.value = shownLocales.value.filter((l) => l !== locale)
  updateLocale(locale, '')
}
</script>

<template>
  <UFormField :label="label" :error="error" :required="required">
    <div class="space-y-2">
      <div v-for="locale in shownLocales" :key="locale" class="flex items-start gap-2">
        <span class="mt-2 w-7 shrink-0 text-xs text-neutral-400">{{ locale.toUpperCase() }}</span>
        <UTextarea
          v-if="multiline"
          :model-value="textFor(locale)"
          :placeholder="label"
          class="w-full"
          @update:model-value="(v) => updateLocale(locale, v)"
        />
        <UInput
          v-else
          :model-value="textFor(locale)"
          :placeholder="label"
          class="w-full"
          @update:model-value="(v) => updateLocale(locale, v)"
        />
        <UButton
          v-if="locale !== requiredLocale"
          icon="i-lucide-x"
          size="xs"
          color="neutral"
          variant="ghost"
          class="mt-1"
          :aria-label="t('common.delete')"
          @click="removeLocale(locale)"
        />
      </div>

      <UDropdownMenu v-if="availableLocales.length" :items="addLocaleItems">
        <UButton icon="i-lucide-plus" size="xs" variant="soft" color="neutral">{{ t('common.addLanguage') }}</UButton>
      </UDropdownMenu>
    </div>
  </UFormField>
</template>
