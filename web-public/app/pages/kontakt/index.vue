<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()

useSeoMeta({
  title: t('seo.contact.title'),
  description: t('seo.contact.description'),
  ogTitle: t('seo.contact.title'),
  ogDescription: t('seo.contact.description'),
  ogImage: '/images/showroom2.png',
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))

const showroomImages = ['/images/showroom1.png', '/images/showroom2.png', '/images/showroom3.png']
const activeImage = ref(0)

function prevImage() {
  activeImage.value = (activeImage.value - 1 + showroomImages.length) % showroomImages.length
}
function nextImage() {
  activeImage.value = (activeImage.value + 1) % showroomImages.length
}

const mapQuery = encodeURIComponent('Veselina Masleše 56, Novi Sad, Srbija')
const mapEmbedSrc = `https://maps.google.com/maps?q=${mapQuery}&t=&z=14&ie=UTF8&iwloc=&output=embed`

const { data: partnersData } = await useAsyncData('public-partners', () => $fetch('/api/partners'))
const partners = computed(() => partnersData.value?.items || [])
</script>

<template>
  <div class="mx-auto max-w-6xl px-4 lg:px-8 py-12 lg:py-16">
    <div class="grid grid-cols-1 lg:grid-cols-5 gap-10 lg:gap-14">
      <div class="lg:col-span-2 lg:order-1">
        <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-10">{{ t('nav.kontakt') }}</h1>

        <h2 class="text-lg font-bold uppercase tracking-wide mb-5">{{ t('contact.showroomLabel') }}</h2>

        <ul class="flex flex-col gap-4 text-black/80">
          <li class="flex items-start gap-3">
            <UIcon name="i-lucide-map-pin" class="w-5 h-5 mt-0.5 text-brand-600 flex-shrink-0" />
            <span class="flex flex-col">
              <span class="font-bold uppercase tracking-wide">{{ t('contact.addressLabel') }}</span>
              <span>{{ t('contact.address') }}</span>
            </span>
          </li>
          <li class="flex items-start gap-3">
            <UIcon name="i-lucide-phone" class="w-5 h-5 mt-0.5 text-brand-600 flex-shrink-0" />
            <span class="flex flex-col">
              <span class="font-bold uppercase tracking-wide">{{ t('contact.phoneLabel') }}</span>
              <a href="tel:+381655632551">+381 65 563 2551</a>
            </span>
          </li>
          <li class="flex items-start gap-3">
            <UIcon name="i-lucide-mail" class="w-5 h-5 mt-0.5 text-brand-600 flex-shrink-0" />
            <span class="flex flex-col">
              <span class="font-bold uppercase tracking-wide">{{ t('contact.emailLabel') }}</span>
              <a href="mailto:serbia@phomi.info">serbia@phomi.info</a>
            </span>
          </li>
        </ul>

        <div class="flex items-center gap-5 mt-8">
          <a href="https://www.instagram.com/phomi_serbia/" target="_blank" rel="noopener" aria-label="Instagram">
            <UIcon name="i-lucide-instagram" class="w-6 h-6 text-black/60 hover:text-brand-600" />
          </a>
          <a href="mailto:serbia@phomi.info" aria-label="Email">
            <UIcon name="i-lucide-mail" class="w-6 h-6 text-black/60 hover:text-brand-600" />
          </a>
        </div>
      </div>

      <div class="lg:col-span-3 lg:order-2 relative">
        <div class="aspect-[4/3] rounded-sm overflow-hidden bg-gray-100">
          <NuxtImg :src="showroomImages[activeImage]" alt="" loading="lazy" class="w-full h-full object-cover" />
        </div>
        <button
          class="absolute top-1/2 left-2 -translate-y-1/2 w-9 h-9 rounded-full bg-white/80 hover:bg-white flex items-center justify-center"
          :aria-label="t('common.prev')"
          @click="prevImage"
        >
          <UIcon name="i-lucide-chevron-left" class="w-5 h-5 text-black" />
        </button>
        <button
          class="absolute top-1/2 right-2 -translate-y-1/2 w-9 h-9 rounded-full bg-white/80 hover:bg-white flex items-center justify-center"
          :aria-label="t('common.next')"
          @click="nextImage"
        >
          <UIcon name="i-lucide-chevron-right" class="w-5 h-5 text-black" />
        </button>
        <div class="flex justify-center gap-2 mt-3">
          <button
            v-for="(img, i) in showroomImages"
            :key="img"
            class="w-2.5 h-2.5 rounded-full"
            :class="i === activeImage ? 'bg-brand-500' : 'bg-black/20'"
            :aria-label="`${i + 1}`"
            @click="activeImage = i"
          />
        </div>
      </div>
    </div>

    <iframe
      :src="mapEmbedSrc"
      class="w-full aspect-[21/9] rounded-sm mt-12 border-0"
      loading="lazy"
      referrerpolicy="no-referrer-when-downgrade"
      :title="t('contact.address')"
    />

    <div class="mt-14">
      <h2 class="text-lg font-bold uppercase tracking-wide mb-5">{{ t('dealer.partnersTitle') }}</h2>
      <p v-if="!partners.length" class="text-black/50">{{ t('dealer.partnersEmpty') }}</p>
      <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="p in partners" :key="p.email" class="border border-black/10 rounded-sm p-4">
          <p class="font-medium">{{ p.name }}</p>
          <p v-if="p.address" class="text-sm text-black/60 mt-1">{{ p.address }}</p>
          <p v-if="p.phone" class="text-sm text-black/60 mt-1"><a :href="`tel:${p.phone}`">{{ p.phone }}</a></p>
          <p v-if="p.email" class="text-sm text-black/60"><a :href="`mailto:${p.email}`">{{ p.email }}</a></p>
        </div>
      </div>
    </div>
  </div>
</template>
