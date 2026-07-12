import { dealerFormSchema } from '~~/server/utils/formSchemas.js'
import { sendNotification } from '~~/server/utils/notificationClient.js'
import { checkRateLimit } from '~~/server/utils/rateLimiter.js'

export default defineEventHandler(async (event) => {
  const ip = getRequestIP(event, { xForwardedFor: true }) || 'unknown'
  if (!checkRateLimit(`dealer-form:${ip}`, { max: 5, windowMs: 60_000 })) {
    throw createError({ statusCode: 429, statusMessage: 'Too many requests' })
  }

  const body = await readBody(event)
  const parsed = dealerFormSchema.safeParse(body)
  if (!parsed.success) {
    throw createError({ statusCode: 400, data: { error: parsed.error.flatten() } })
  }
  if (parsed.data.website) {
    // Honeypot tripped — pretend success, do nothing.
    return { ok: true }
  }

  const { companyName, contactName, phone, email, city, message } = parsed.data
  const lines = [`Company: ${companyName}`, `Contact: ${contactName}`]
  if (phone) lines.push(`Phone: ${phone}`)
  lines.push(`Email: ${email}`, `City: ${city}`)
  if (message) lines.push('', message)

  await sendNotification({
    subject: 'Website dealer application',
    message: lines.join('\n'),
    email,
  })
  return { ok: true }
})
