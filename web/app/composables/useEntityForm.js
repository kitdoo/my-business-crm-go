import { getEntityConfig } from '~/config/entities'

// Backs <EntityForm> (TD §9.1/§9.2): loads a record (edit mode) or starts
// blank (create mode), tracks the original snapshot to diff into a
// FieldMask on save, and always carries the etag through update/delete so
// individual forms never have to remember to do it themselves.
export function useEntityForm(entityKey, id) {
  const config = getEntityConfig(entityKey)
  const api = useEntityApi(entityKey)
  const { handle } = useApiErrorHandler()

  const isCreate = !id
  const form = ref(blankForm(config))
  const original = ref(null)
  const etag = ref(null)
  const loading = ref(false)
  const saving = ref(false)
  const fieldErrors = ref({})
  const etagConflict = ref(false)
  // Full record as last loaded from the server, unlike `form` which only
  // tracks config.form.fields — needed by things like a Warehouse
  // Deactivate action button whose visibility depends on `status`, a
  // field intentionally absent from the editable form (TD §12.2).
  const record = ref(null)

  function blankForm(cfg) {
    const state = {}
    for (const field of cfg.form.fields) {
      if (field.editOnly) continue
      state[field.key] = field.type === 'localizedString' ? { values: {} } : null
    }
    return state
  }

  function applyRecord(rec) {
    const state = {}
    for (const field of config.form.fields) {
      state[field.key] = rec[field.key] ?? (field.type === 'localizedString' ? { values: {} } : null)
    }
    form.value = state
    original.value = state
    etag.value = rec.etag
    record.value = rec
  }

  async function load() {
    if (isCreate) return
    loading.value = true
    try {
      const record = await api.get(id)
      applyRecord(record)
    } catch (err) {
      handle(err)
    } finally {
      loading.value = false
    }
  }

  async function save() {
    fieldErrors.value = {}
    etagConflict.value = false
    saving.value = true
    try {
      if (isCreate) {
        const fields = {}
        for (const field of config.form.fields) {
          if (field.editOnly) continue
          fields[field.key] = form.value[field.key]
        }
        const record = await api.create(fields)
        return record
      }

      const editableKeys = config.form.fields.map((f) => f.key)
      const mask = buildUpdateMask(original.value, form.value, editableKeys)
      if (mask.length === 0) return original.value

      const fields = {}
      for (const key of mask) fields[key] = form.value[key]

      const record = await api.update(id, fields, mask, etag.value)
      applyRecord(record)
      return record
    } catch (err) {
      handle(err, {
        onFieldError: (field, message) => {
          fieldErrors.value = { ...fieldErrors.value, [field]: message }
        },
        onEtagConflict: () => {
          etagConflict.value = true
        },
      })
      throw err
    } finally {
      saving.value = false
    }
  }

  async function reloadAfterConflict() {
    etagConflict.value = false
    await load()
  }

  return {
    form,
    record,
    etag,
    loading,
    saving,
    fieldErrors,
    etagConflict,
    isCreate,
    load,
    save,
    reloadAfterConflict,
  }
}
