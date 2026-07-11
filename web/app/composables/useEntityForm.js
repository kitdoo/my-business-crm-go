import { getEntityConfig } from '~/config/entities'

// Backs <EntityForm> (TD §9.1/§9.2): loads a record (edit mode) or starts
// blank (create mode), tracks the original snapshot to diff into a
// FieldMask on save, and always carries the etag through update/delete so
// individual forms never have to remember to do it themselves.
export function useEntityForm(entityKey, id, initialValues = {}) {
  const config = getEntityConfig(entityKey)
  const api = useEntityApi(entityKey)
  const { handle } = useApiErrorHandler()

  const isCreate = !id
  const form = ref({ ...blankForm(config), ...(isCreate ? initialValues : {}) })
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

  // Per-type blank value, so a field's shape is always what its input
  // component expects — never bare `null` for something a v-model expects
  // to be an array/object (relationMulti/images want [], keyValueLocalized
  // wants {}), or the input has to null-guard on every read.
  function blankValue(field) {
    if (field.type === 'localizedString') return { values: {} }
    if (field.type === 'relationMulti' || field.type === 'images') return []
    if (field.type === 'keyValueLocalized') return {}
    return null
  }

  function blankForm(cfg) {
    const state = {}
    for (const field of cfg.form.fields) {
      if (field.editOnly) continue
      state[field.key] = blankValue(field)
    }
    return state
  }

  function applyRecord(rec) {
    const state = {}
    for (const field of config.form.fields) {
      // createOnly fields (e.g. User.password) don't exist on the fetched
      // record and aren't editable — never part of the update FieldMask.
      if (field.createOnly) continue
      state[field.key] = rec[field.key] ?? blankValue(field)
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

      const editableKeys = config.form.fields.filter((f) => !f.createOnly).map((f) => f.key)
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
