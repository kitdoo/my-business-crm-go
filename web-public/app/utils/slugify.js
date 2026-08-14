// Category URLs (see katalog/kategorija/[category].vue) are keyed on a slug
// derived from the category's Serbian name — there's no dedicated slug field
// on the backend yet, so this is the frontend's own deterministic mapping.
// Serbian (Latin) digraphs are transliterated explicitly since NFKD doesn't
// decompose đ; everything else goes through NFKD + combining-mark strip to
// catch stray accents. Serbian is officially written in both scripts, so a
// category name entered in Cyrillic (e.g. "Опека") needs the same treatment
// — without it, a Cyrillic-only name has no [a-z0-9] characters left after
// stripping, collapses to an empty slug, and building the category URL from
// that empty param throws (see CatalogPage.vue's categorySlugOf, which
// falls back to the category id when this returns ''). Mirrored in
// server/utils/slugify.js (Nitro routes can't import from app/) — keep both
// in sync if this changes.
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
