// Mirror of app/utils/slugify.js — Nitro routes (sitemap.xml.js) can't
// import from app/, so this is duplicated rather than shared. Keep both in
// sync if this changes; see the app/ copy for the full rationale.
const DIGRAPHS = { č: 'c', ć: 'c', đ: 'dj', š: 's', ž: 'z' }
const CYRILLIC = {
  а: 'a', б: 'b', в: 'v', г: 'g', д: 'd', ђ: 'dj', е: 'e', ж: 'z', з: 'z', и: 'i',
  ј: 'j', к: 'k', л: 'l', љ: 'lj', м: 'm', н: 'n', њ: 'nj', о: 'o', п: 'p', р: 'r',
  с: 's', т: 't', ћ: 'c', у: 'u', ф: 'f', х: 'h', ц: 'c', ч: 'c', џ: 'dz', ш: 's',
}

export function slugify(text) {
  return String(text || '')
    .toLowerCase()
    .replace(/[čćđšž]/g, (ch) => DIGRAPHS[ch])
    .replace(/[а-шђјљњћџ]/g, (ch) => CYRILLIC[ch] ?? ch)
    .normalize('NFKD')
    .replace(/[̀-ͯ]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}
