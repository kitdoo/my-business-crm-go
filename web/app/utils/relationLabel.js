import { localizedText } from '~/utils/localizedText.js'

// Shared "how do I show this related record as text" rule, used by
// RelationLabel/RelationSelect/EntityFilterBar's relation-option loader —
// one place instead of three copies. Most entities have a `name`
// (localized or plain); Sale doesn't (TD §12.3, no name field at all) —
// falls back to its human-readable Number instead of a raw UUID.
// ProductVariant has neither — sku is its human-readable identifier.
export function relationLabel(item, locale) {
  if (item.name) {
    return typeof item.name === 'object' ? localizedText(item.name, locale) : item.name
  }
  if (item.number != null) return `#${item.number}`
  if (item.sku) return item.sku
  return item.id
}
