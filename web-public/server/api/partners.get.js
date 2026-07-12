import { listPublicPartners } from '~~/server/utils/partnersClient.js'

// GET /api/partners — active dealer/partner list for the "Postani diler"
// page (item 13), sourced live from the CRM admin panel.
export default defineEventHandler(async () => {
  const items = await listPublicPartners()
  return {
    items: items.map((p) => ({
      name: p.name,
      phone: p.phone,
      email: p.email,
      address: p.address,
    })),
  }
})
