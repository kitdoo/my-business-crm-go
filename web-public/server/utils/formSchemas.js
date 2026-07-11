import { z } from 'zod'

// TZ §7.2 / §7.3. `website` is a honeypot field — real visitors never see
// or fill it (hidden via CSS in the form component); any non-empty value
// means a bot filled every field it could find.
export const dealerFormSchema = z.object({
  companyName: z.string().min(1).max(200),
  contactName: z.string().min(1).max(200),
  phone: z.string().max(32).optional().default(''),
  email: z.string().email(),
  city: z.string().max(200),
  message: z.string().max(2000).optional().default(''),
  website: z.string().max(0).optional().default(''),
})

export const contactFormSchema = z.object({
  name: z.string().min(1).max(200),
  email: z.string().email(),
  phone: z.string().max(32).optional().default(''),
  message: z.string().min(1).max(2000),
  website: z.string().max(0).optional().default(''),
})
