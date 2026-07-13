<script setup>
// Generic filter bar (TD §9.3) rendered above every EntityDataTable from
// config.list.filters — one row, no per-entity filter UI needed. Supports
// the filter `type`s entity configs already declare: multiselect/select
// (options from an enum or a relation's full list), number, numberRange
// (one compact min–max control, e.g. Inventory.quantity), and
// periodFilter (reuses <DateRangePicker> — just the from/to range piece
// of the dashboard's PeriodFilter, no today/7d/month presets here).
import { getEnumOptions } from '~/config/enums.js'
import { relationLabel } from '~/utils/relationLabel.js'

const props = defineProps({
  filters: { type: Array, required: true },
  modelValue: { type: Object, required: true },
})
const emit = defineEmits(['update:modelValue'])

const { t, locale } = useI18n()
const relationOptions = ref({})

async function loadRelationOptions(entityKey) {
  if (relationOptions.value[entityKey]) return
  const api = useEntityApi(entityKey)
  const res = await api.list({ pagination: { limit: 200 } })
  relationOptions.value = {
    ...relationOptions.value,
    [entityKey]: res.items.map((item) => ({ label: relationLabel(item, locale.value), value: item.id })),
  }
}

function optionsFor(filter) {
  if (!filter.optionsFrom) return []
  const [kind, name] = filter.optionsFrom.split(':')
  if (kind === 'enum') {
    return getEnumOptions(name).map((v) => ({ label: t(`enums.status.${v}`), value: v }))
  }
  if (kind === 'relation') {
    return relationOptions.value[name] || []
  }
  return []
}

// Client-only: triggering these $fetch calls from inside a template
// getter (as part of SSR render) used plain $fetch, which doesn't carry
// the incoming request's session cookie during SSR (unlike
// useRequestFetch — see stores/auth.js) and 401'd on every server
// render. Loading once on mount sidesteps that entirely — the dropdown
// options aren't needed for first paint anyway.
onMounted(() => {
  for (const filterDef of props.filters) {
    if (!filterDef.optionsFrom) continue
    const [kind, name] = filterDef.optionsFrom.split(':')
    if (kind === 'relation') loadRelationOptions(name)
  }
})

function update(key, value) {
  emit('update:modelValue', { ...props.modelValue, [key]: value })
}
</script>

<template>
  <div class="flex flex-wrap items-end gap-3">
    <template v-for="filterDef in filters" :key="filterDef.key">
      <UFormField v-if="filterDef.type === 'multiselect'" :label="t(filterDef.label)" class="w-48">
        <USelectMenu
          :model-value="modelValue[filterDef.key] || []"
          :items="optionsFor(filterDef)"
          multiple
          value-key="value"
          class="w-full"
          @update:model-value="(v) => update(filterDef.key, v)"
        />
      </UFormField>

      <UFormField v-else-if="filterDef.type === 'select'" :label="t(filterDef.label)" class="w-48">
        <USelectMenu
          :model-value="modelValue[filterDef.key] ?? null"
          :items="optionsFor(filterDef)"
          value-key="value"
          class="w-full"
          @update:model-value="(v) => update(filterDef.key, v)"
        />
      </UFormField>

      <UFormField v-else-if="filterDef.type === 'number'" :label="t(filterDef.label)" class="w-32">
        <UInputNumber
          :model-value="modelValue[filterDef.key] ?? null"
          class="w-full"
          @update:model-value="(v) => update(filterDef.key, v)"
        />
      </UFormField>

      <UFormField v-else-if="filterDef.type === 'periodFilter'" :label="t(filterDef.label)">
        <DateRangePicker
          :model-value="modelValue[filterDef.key] || {}"
          :active="!!modelValue[filterDef.key]"
          @update:model-value="(v) => update(filterDef.key, v)"
        />
      </UFormField>

      <UFormField v-else-if="filterDef.type === 'numberRange'" :label="t(filterDef.label)">
        <div class="flex items-center gap-1">
          <UInputNumber
            :model-value="modelValue[filterDef.key]?.min ?? null"
            :placeholder="t('common.min')"
            class="w-24"
            @update:model-value="(v) => update(filterDef.key, { ...modelValue[filterDef.key], min: v })"
          />
          <span class="text-neutral-400">–</span>
          <UInputNumber
            :model-value="modelValue[filterDef.key]?.max ?? null"
            :placeholder="t('common.max')"
            class="w-24"
            @update:model-value="(v) => update(filterDef.key, { ...modelValue[filterDef.key], max: v })"
          />
        </div>
      </UFormField>
    </template>
  </div>
</template>
