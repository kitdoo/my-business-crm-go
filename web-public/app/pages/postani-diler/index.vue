<script setup>
const { t } = useI18n()
const localeHead = useLocaleHead()
const localePath = useLocalePath()
const router = useRouter()

useSeoMeta({
  title: t('seo.dealer.title'),
  description: t('seo.dealer.description'),
})
useHead(() => ({ link: localeHead.value.link, meta: localeHead.value.meta }))

function onSubmitted() {
  router.push(localePath('/postani-diler/hvala'))
}
</script>

<template>
  <div class="mx-auto max-w-2xl px-4 lg:px-8 py-12 lg:py-16">
    <h1 class="text-2xl lg:text-3xl font-bold uppercase tracking-wide mb-4">{{ t('nav.postaniDiler') }}</h1>
    <p class="text-black/70 mb-10">{{ t('dealer.intro') }}</p>
    <DealerForm @submitted="onSubmitted" />
  </div>
</template>
