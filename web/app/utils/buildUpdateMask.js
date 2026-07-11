// Compares the record as loaded (`original`) against the current form
// state (`current`) and returns only the top-level field paths that
// actually changed — the FieldMask sent on update (TD §9.2). Every form
// goes through this so "send everything always" never creeps back in.
export function buildUpdateMask(original, current, fieldKeys) {
  const mask = []
  for (const key of fieldKeys) {
    const before = original?.[key]
    const after = current?.[key]
    if (JSON.stringify(before) !== JSON.stringify(after)) {
      mask.push(key)
    }
  }
  return mask
}
