// og:image (and similar crawler-facing meta) must be an absolute URL per
// the Open Graph spec — a relative path renders fine in-app but social
// crawlers (Facebook, Telegram, VK...) that don't resolve it against the
// page URL simply fail to fetch the preview image.
export function useAbsoluteUrl(path) {
  const { siteUrl } = useRuntimeConfig().public
  return `${siteUrl}${path}`
}
