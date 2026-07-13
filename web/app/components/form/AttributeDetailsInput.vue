<script setup>
// Editor for Product.details (characteristics), replacing the old
// free-form KeyValueLocalizedInput: the set of possible keys comes from
// the backend-defined catalog (ProductAttributeDefinition — read-only,
// no admin CRUD UI, seeded by a developer), each carrying a localized
// label and a public/private flag.
//
// Details are optional (TD §12.1): a product doesn't need any
// characteristic filled in, so nothing here is rendered until the
// operator explicitly adds it via "+ add characteristic", picking from
// the catalog's definitions not already added. Each added row can be
// removed again, dropping its key entirely rather than leaving a blank
// value behind.
//
// Any key already in modelValue that doesn't match a known definition
// (legacy data, or a definition removed since) is preserved untouched and
// merged back in on emit — just not rendered as its own row.
const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  locales: { type: Array, required: true },
  label: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()
const { locale } = useI18n()

const api = useEntityApi('productAttributeDefinitions')
const loading = ref(true)
const definitions = ref([])

async function loadDefinitions() {
  loading.value = true
  try {
    const res = await api.list({ pagination: { limit: 200 } })
    definitions.value = res.items || []
  } finally {
    loading.value = false
  }
}
onMounted(loadDefinitions)

function definitionLabel(def) {
  return def.label?.values?.[locale.value] || def.label?.values?.sr || def.key
}

const definitionByKey = computed(() => new Map(definitions.value.map((def) => [def.key, def])))

const addedDefinitions = computed(() =>
  definitions.value.filter((def) => Object.prototype.hasOwnProperty.call(props.modelValue || {}, def.key)),
)
const availableDefinitions = computed(() =>
  definitions.value.filter((def) => !Object.prototype.hasOwnProperty.call(props.modelValue || {}, def.key)),
)
const addItems = computed(() => [
  availableDefinitions.value.map((def) => ({ label: definitionLabel(def), onSelect: () => addCharacteristic(def.key) })),
])

function valueFor(key) {
  return props.modelValue?.[key] || { values: {} }
}

function updateValue(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}

function addCharacteristic(key) {
  if (Object.prototype.hasOwnProperty.call(props.modelValue || {}, key)) return
  emit('update:modelValue', { ...props.modelValue, [key]: { values: {} } })
}

function removeCharacteristic(key) {
  const next = { ...props.modelValue }
  delete next[key]
  emit('update:modelValue', next)
}
</script>

<template>
  <div class="space-y-3 md:col-span-1">
    <span class="block text-sm font-medium text-highlighted">{{ label }}</span>
    <div v-if="loading" class="py-4 text-center text-sm text-neutral-500">{{ t('common.loading') }}</div>
    <template v-else>
      <p v-if="!addedDefinitions.length" class="text-sm text-neutral-500">{{ t('common.empty') }}</p>
      <div v-else class="space-y-3">
        <div v-for="def in addedDefinitions" :key="def.id" class="rounded-md border border-default p-3">
          <div class="mb-1 flex items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <span class="text-xs font-medium text-neutral-500">{{ definitionLabel(def) }}</span>
              <UBadge :color="def.isPublic ? 'success' : 'neutral'" variant="subtle" size="sm">
                {{ def.isPublic ? t('fields.attributePublic') : t('fields.attributePrivate') }}
              </UBadge>
            </div>
            <UButton
              icon="i-lucide-x"
              size="xs"
              color="neutral"
              variant="ghost"
              :aria-label="t('common.delete')"
              @click="removeCharacteristic(def.key)"
            />
          </div>
          <LocalizedStringInput
            :model-value="valueFor(def.key)"
            :locales="locales"
            required-locale="sr"
            @update:model-value="(v) => updateValue(def.key, v)"
          />
        </div>
      </div>

      <UDropdownMenu v-if="availableDefinitions.length" :items="addItems">
        <UButton icon="i-lucide-plus" size="xs" variant="soft" color="neutral">
          {{ t('common.addCharacteristic') }}
        </UButton>
      </UDropdownMenu>
    </template>
  </div>
</template>
