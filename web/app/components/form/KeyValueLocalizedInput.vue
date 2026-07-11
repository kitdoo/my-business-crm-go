<script setup>
// Editor for `map<string, LocalizedString>` fields (TD §9.1 — Product.details:
// arbitrary characteristics like color/size, key -> localized value).
// Keeps a local list of {id, key, value} entries for stable editing —
// renaming a key would otherwise reorder/lose focus if driven straight off
// object keys. Safe without resyncing to external modelValue changes
// because <EntityForm> remounts this whole tree per record
// (:key="editingId ?? 'create'"), so there's no case where modelValue
// changes out from under an already-mounted instance.
const props = defineProps({
  modelValue: { type: Object, default: () => ({}) },
  locales: { type: Array, required: true },
  label: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue'])
const { t } = useI18n()

let nextId = 0
const entries = ref(
  Object.entries(props.modelValue || {}).map(([key, value]) => ({
    id: nextId++,
    key,
    value: value || { values: {} },
  })),
)

function sync() {
  const obj = {}
  for (const entry of entries.value) {
    const key = entry.key.trim()
    if (key) obj[key] = entry.value
  }
  emit('update:modelValue', obj)
}

function addEntry() {
  entries.value = [...entries.value, { id: nextId++, key: '', value: { values: {} } }]
}

function removeEntry(id) {
  entries.value = entries.value.filter((entry) => entry.id !== id)
  sync()
}
</script>

<template>
  <div class="space-y-2 md:col-span-2">
    <span class="block text-sm font-medium text-highlighted">{{ label }}</span>
    <div v-for="entry in entries" :key="entry.id" class="flex gap-2 items-start rounded-md border border-default p-3">
      <UInput
        v-model="entry.key"
        :placeholder="t('fields.detailKey')"
        class="w-40 shrink-0"
        @update:model-value="sync"
      />
      <LocalizedStringInput
        v-model="entry.value"
        :locales="locales"
        required-locale="sr"
        :label="t('fields.detailValue')"
        class="flex-1"
        @update:model-value="sync"
      />
      <UButton icon="i-lucide-x" color="error" variant="ghost" class="mt-1" @click="removeEntry(entry.id)" />
    </div>
    <UButton icon="i-lucide-plus" variant="soft" size="sm" @click="addEntry">{{ t('common.add') }}</UButton>
  </div>
</template>
