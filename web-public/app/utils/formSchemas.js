import { z } from 'zod'

// Client-side mirror of server/utils/formSchemas.js — the server route
// re-validates independently and is the source of truth; this copy is
// only for inline form UX (TZ §7.1).
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
