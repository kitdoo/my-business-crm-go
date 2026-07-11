import { contactFormSchema } from '~~/server/utils/formSchemas.js'
import { submitContactForm } from '~~/server/utils/submitForm.js'
import { checkRateLimit } from '~~/server/utils/rateLimiter.js'

export default defineEventHandler(async (event) => {
  const ip = getRequestIP(event, { xForwardedFor: true }) || 'unknown'
  if (!checkRateLimit(`contact-form:${ip}`, { max: 5, windowMs: 60_000 })) {
    throw createError({ statusCode: 429, statusMessage: 'Too many requests' })
  }

  const body = await readBody(event)
  const parsed = contactFormSchema.safeParse(body)
  if (!parsed.success) {
    throw createError({ statusCode: 400, data: { error: parsed.error.flatten() } })
  }
  if (parsed.data.website) {
    return { ok: true }
  }

  const { website, ...payload } = parsed.data
  return submitContactForm(payload)
})
