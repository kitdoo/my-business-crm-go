export default defineEventHandler((event) => {
  const config = useRuntimeConfig()
  const siteUrl = config.public.siteUrl || ''
  setHeader(event, 'Content-Type', 'text/plain')
  return [
    'User-agent: *',
    'Allow: /',
    siteUrl ? `Sitemap: ${siteUrl}/sitemap.xml` : '',
  ].filter(Boolean).join('\n')
})
