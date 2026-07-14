// Static content config for AdvantagesSection ("Ključne prednosti", TZ
// §4.3) — not CRM data, so it lives here as i18n-keyed config rather than
// a backend entity. 6 cards, 3 columns x 2 rows. Icons are line-art PNGs
// (some white-stroke, some black-stroke) — AdvantagesSection.vue forces
// them all to solid white via [filter:brightness(0)_invert(1)] so the
// source color doesn't matter.
export const advantages = [
  { icon: '/images/icons/advantages/1.png', titleKey: 'home.advantages.versatility.title', textKey: 'home.advantages.versatility.text' },
  { icon: '/images/icons/advantages/2.png', titleKey: 'home.advantages.practical.title', textKey: 'home.advantages.practical.text' },
  { icon: '/images/icons/advantages/3.png', titleKey: 'home.advantages.designs.title', textKey: 'home.advantages.designs.text' },
  { icon: '/images/icons/advantages/4.png', titleKey: 'home.advantages.resistance.title', textKey: 'home.advantages.resistance.text' },
  { icon: '/images/icons/advantages/5.png', titleKey: 'home.advantages.eco.title', textKey: 'home.advantages.eco.text' },
  { icon: '/images/icons/advantages/6.png', titleKey: 'home.advantages.install.title', textKey: 'home.advantages.install.text' },
]
