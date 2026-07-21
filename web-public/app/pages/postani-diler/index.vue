<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()
const localePath = useLocalePath()
const router = useRouter()

const { data: partnersData } = await useAsyncData('public-partners', () => $fetch('/api/partners'))
const partners = computed(() => partnersData.value?.items || [])

useSeoMeta({
  title: t('seo.dealer.title'),
  description: t('seo.dealer.description'),
  ogTitle: t('seo.dealer.title'),
  ogDescription: t('seo.dealer.description'),
  ogImage: '/images/social-share.jpg',
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))

function onSubmitted() {
  router.push(localePath('/postani-diler/hvala'))
}
</script>

<template>
  <div class="mx-auto max-w-7xl 2xl:max-w-[1800px] px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-10">{{ t('nav.postaniDiler') }}</h1>

    <div class="grid grid-cols-1 lg:grid-cols-5 gap-10 lg:gap-14">
      <div class="lg:col-span-3">
        <h2 class="text-lg font-bold uppercase tracking-wide mb-3">{{ t('dealer.whyTitle') }}</h2>
        <p class="text-black/70 leading-relaxed mb-2">{{ t('dealer.whyText1') }}</p>
        <p class="text-black/70 leading-relaxed mb-8">{{ t('dealer.whyText2') }}</p>

        <p class="text-black/70 mb-6">{{ t('dealer.intro') }}</p>
        <div class="w-full bg-gray-100 rounded-sm p-6 lg:p-8">
          <DealerForm @submitted="onSubmitted" />
        </div>
      </div>

      <div class="lg:col-span-2">
        <h2 class="text-lg font-bold uppercase tracking-wide mb-4">{{ t('dealer.partnersTitle') }}</h2>
        <p v-if="!partners.length" class="text-black/50">{{ t('dealer.partnersEmpty') }}</p>
        <ul v-else class="flex flex-col gap-4">
          <li v-for="p in partners" :key="p.email" class="border border-black/10 rounded-sm p-4">
            <p class="font-medium">{{ p.name }}</p>
            <p v-if="p.address" class="text-sm text-black/60 mt-1">{{ p.address }}</p>
            <p v-if="p.phone" class="text-sm text-black/60 mt-1"><a :href="`tel:${p.phone}`">{{ p.phone }}</a></p>
            <p v-if="p.email" class="text-sm text-black/60"><a :href="`mailto:${p.email}`">{{ p.email }}</a></p>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
