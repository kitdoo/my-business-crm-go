<script setup>
const { t, locale } = useI18n()
const localeHead = useSeoHead()

useSeoMeta({
  title: t('seo.home.title'),
  description: t('seo.home.description'),
  ogTitle: t('seo.home.title'),
  ogDescription: t('seo.home.description'),
  ogImage: useAbsoluteUrl('/images/social-share.jpg'),
})
useHead(() => ({
  htmlAttrs: { lang: locale.value },
  link: localeHead.value.link,
  meta: localeHead.value.meta,
}))
</script>

<template>
  <div>
    <HeroSection />
    <AboutSection />
    <ProjectsPreviewSection />
    <DezenStatsSection />
    <WhyPhomiSection />
    <AdvantagesSection />
    <InstallSection />
    <WorldPresenceSection />
    <BecomePartnerSection />
  </div>
</template>
