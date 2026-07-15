<script setup>
// Public, binary version of the admin StatusBadge (TZ §8.3) — visitors only
// ever see "Dostupno" / "Nije dostupno". Driven by stock (available =
// in-stock), not the admin Draft/Active/Archived status — every product and
// variant reaching web-public is already forced ACTIVE server-side (see
// catalogClient.js), so an admin-status-based badge would never vary; stock
// is the only thing a visitor actually needs signaled here.
import { tokens } from '~/design/tokens.js'

const props = defineProps({
  available: { type: Boolean, required: true },
})

const family = computed(() => (props.available ? 'active' : 'inactive'))
const colors = computed(() => tokens.color.status[family.value])

const { t } = useI18n()
const label = computed(() => (props.available ? t('catalog.available') : t('catalog.unavailable')))
</script>

<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
    :style="{ backgroundColor: colors.bg, color: colors.text }"
  >
    {{ label }}
  </span>
</template>
