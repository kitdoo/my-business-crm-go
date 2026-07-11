// The one place the current-locale -> sr fallback rule lives (TD §5.3) —
// used by both <LocalizedText> (read-only display) and anywhere a plain
// string label is needed (e.g. <RelationSelect> option labels).
export function localizedText(value, locale) {
  const values = value?.values || {}
  return values[locale] ?? values.sr ?? ''
}
