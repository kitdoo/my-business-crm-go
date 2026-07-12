import { contactFormSchema } from '~~/server/utils/formSchemas.js'
import { sendNotification } from '~~/server/utils/notificationClient.js'
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
    // Honeypot tripped — pretend success, do nothing.
    return { ok: true }
  }

  const { name, email, phone, message } = parsed.data
  const lines = [`Name: ${name}`, `Email: ${email}`]
  if (phone) lines.push(`Phone: ${phone}`)
  lines.push('', message)

  await sendNotification({
    subject: 'Website contact form submission',
    message: lines.join('\n'),
    email,
  })
  return { ok: true }
})
