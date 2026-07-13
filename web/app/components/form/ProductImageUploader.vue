<script setup>
// Upload/reorder/remove for Product.imageIds (TD §4.6/§12.6). v-model is
// the ordered array of ids — display order is significant ("first is
// main"), matching the backend's "full replacement, in display order"
// contract for Update. Removing an image only drops its id from the
// array; the file itself is never deleted (imageIds only ever
// references already-uploaded files, doesn't manage their lifetime).
const props = defineProps({
  modelValue: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const toast = useToast()
const runtimeConfig = useRuntimeConfig()
const uploading = ref(false)
const fileInput = ref(null)

// Sniffed by content (magic bytes), not filename/extension or the
// browser-reported File.type — the backend does the same, this is just
// UX so a bad file doesn't get uploaded before being rejected.
const MAGIC = [
  { mime: 'image/jpeg', bytes: [0xff, 0xd8, 0xff] },
  { mime: 'image/png', bytes: [0x89, 0x50, 0x4e, 0x47] },
  { mime: 'image/gif', bytes: [0x47, 0x49, 0x46, 0x38] },
]

function sniffMime(bytes) {
  for (const { mime, bytes: sig } of MAGIC) {
    if (sig.every((b, i) => bytes[i] === b)) return mime
  }
  // WEBP: "RIFF" .... "WEBP"
  if (
    bytes[0] === 0x52 &&
    bytes[1] === 0x49 &&
    bytes[2] === 0x46 &&
    bytes[3] === 0x46 &&
    bytes[8] === 0x57 &&
    bytes[9] === 0x45 &&
    bytes[10] === 0x42 &&
    bytes[11] === 0x50
  ) {
    return 'image/webp'
  }
  return null
}

async function readHeader(file) {
  const buf = await file.slice(0, 12).arrayBuffer()
  return new Uint8Array(buf)
}

function onPick() {
  fileInput.value?.click()
}

async function onFileChange(e) {
  const file = e.target.files?.[0]
  e.target.value = ''
  if (!file) return

  if (file.size > runtimeConfig.public.imagesMaxSizeBytes) {
    toast.add({ title: t('products.imageTooLarge'), color: 'error' })
    return
  }
  const mime = sniffMime(await readHeader(file))
  if (!mime) {
    toast.add({ title: t('products.imageUnsupportedType'), color: 'error' })
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    // Backend wraps successful responses in a {"data": ...} envelope
    // (go-atlas httpserver/writer.Default) — the proxy passes it through
    // unchanged, so the id is nested under .data.id, not top-level.
    const { data } = await $fetch('/api/images/upload', { method: 'POST', body: formData })
    emit('update:modelValue', [...props.modelValue, data.id])
  } catch (err) {
    toast.add({ title: err?.data?.error?.message || t('errors.generic'), color: 'error' })
  } finally {
    uploading.value = false
  }
}

function removeImage(id) {
  emit(
    'update:modelValue',
    props.modelValue.filter((imgId) => imgId !== id),
  )
}

function moveImage(index, direction) {
  const target = index + direction
  if (target < 0 || target >= props.modelValue.length) return
  const next = [...props.modelValue]
  ;[next[index], next[target]] = [next[target], next[index]]
  emit('update:modelValue', next)
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex flex-wrap gap-3">
      <div
        v-for="(id, index) in modelValue"
        :key="id"
        class="relative w-24 h-24 rounded-md border border-default overflow-hidden group"
      >
        <img :src="`/api/images/${id}`" class="w-full h-full object-cover" :alt="`image ${index + 1}`" />
        <div
          class="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 flex items-center justify-center gap-1 transition-opacity"
        >
          <UButton
            icon="i-lucide-chevron-left"
            size="xs"
            color="neutral"
            :disabled="index === 0"
            @click="moveImage(index, -1)"
          />
          <UButton icon="i-lucide-x" size="xs" color="error" @click="removeImage(id)" />
          <UButton
            icon="i-lucide-chevron-right"
            size="xs"
            color="neutral"
            :disabled="index === modelValue.length - 1"
            @click="moveImage(index, 1)"
          />
        </div>
        <span v-if="index === 0" class="absolute top-1 left-1 rounded bg-brand-500 text-white text-[10px] px-1">
          {{ t('products.mainImage') }}
        </span>
      </div>

      <button
        type="button"
        class="w-24 h-24 rounded-md border-2 border-dashed border-neutral-300 flex items-center justify-center text-neutral-400 hover:border-brand-500 hover:text-brand-500 disabled:opacity-50"
        :disabled="uploading"
        @click="onPick"
      >
        <UIcon v-if="!uploading" name="i-lucide-plus" class="size-6" />
        <UIcon v-else name="i-lucide-loader-2" class="size-6 animate-spin" />
      </button>
    </div>
    <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onFileChange" />
  </div>
</template>
