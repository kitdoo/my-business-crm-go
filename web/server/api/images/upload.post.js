import { requireSession } from '~~/server/utils/session'

// Proxies POST /images on the backend HTTP host (not gRPC, TD §4.6) —
// admin-only there (checked via the same bearer token as gRPC), so the
// browser never gets a direct route to it: same-origin only. Preflight
// MIME/size validation happens client-side in ProductImageUploader for UX;
// the backend re-validates by content regardless.
export default defineEventHandler(async (event) => {
  const session = requireSession(event)
  const config = useRuntimeConfig()

  const parts = await readMultipartFormData(event)
  const filePart = parts?.find((p) => p.name === 'file')
  if (!filePart) {
    throw createError({
      statusCode: 400,
      data: { error: { code: 'INVALID_ARGUMENT', message: 'missing "file" field' } },
    })
  }

  const formData = new FormData()
  formData.append('file', new Blob([filePart.data], { type: filePart.type }), filePart.filename || 'upload')

  try {
    return await $fetch(`${config.images.baseUrl}/images`, {
      method: 'POST',
      headers: { authorization: `Bearer ${session.token}` },
      body: formData,
    })
  } catch (err) {
    const status = err?.response?.status || err?.statusCode || 502
    const CODE_BY_STATUS = {
      400: 'INVALID_ARGUMENT',
      401: 'UNAUTHENTICATED',
      403: 'PERMISSION_DENIED',
      413: 'FAILED_PRECONDITION',
      415: 'FAILED_PRECONDITION',
    }
    throw createError({
      statusCode: status,
      data: { error: { code: CODE_BY_STATUS[status] || 'INTERNAL', message: 'Image upload failed' } },
    })
  }
})
