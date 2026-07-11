<script setup>
// Public, binary version of the admin StatusBadge (TZ §8.3) — visitors only
// ever see "Dostupno" / "Nije dostupno", never Draft/Archived internals.
import { tokens, CATALOG_STATUS_MAP } from '~/design/tokens.js'

const props = defineProps({
  status: { type: String, required: true },
})

const family = computed(() => CATALOG_STATUS_MAP[props.status] || 'inactive')
const colors = computed(() => tokens.color.status[family.value])

const { t } = useI18n()
const label = computed(() => (family.value === 'active' ? t('catalog.available') : t('catalog.unavailable')))
</script>

<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
    :style="{ backgroundColor: colors.bg, color: colors.text }"
  >
    {{ label }}
  </span>
</template>
