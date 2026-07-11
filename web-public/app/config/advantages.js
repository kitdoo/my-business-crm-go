// Static content config for AdvantagesSection ("Ključne prednosti", TZ
// §4.3) — not CRM data, so it lives here as i18n-keyed config rather than
// a backend entity. 6 cards, 3 columns x 2 rows.
// mini_clock.jpg is not uploaded yet (to be added later) — the <img> will
// 404 until then, which is expected.
export const advantages = [
  { icon: '/images/mini_house.jpg', titleKey: 'home.advantages.versatility.title', textKey: 'home.advantages.versatility.text' },
  { icon: '/images/mini_hand.jpg', titleKey: 'home.advantages.practical.title', textKey: 'home.advantages.practical.text' },
  { icon: '/images/mini_paint.jpg', titleKey: 'home.advantages.designs.title', textKey: 'home.advantages.designs.text' },
  { icon: '/images/mini_fire.jpg', titleKey: 'home.advantages.resistance.title', textKey: 'home.advantages.resistance.text' },
  { icon: '/images/mini_wood.jpg', titleKey: 'home.advantages.eco.title', textKey: 'home.advantages.eco.text' },
  { icon: '/images/mini_clock.jpg', titleKey: 'home.advantages.install.title', textKey: 'home.advantages.install.text' },
]
