// Proto enum values as returned by the server (proto-loader enums:'String'
// mode gives us the full name, e.g. 'BRAND_STATUS_ACTIVE'). *_UNSPECIFIED
// is never offered as a pickable option (TD §11).
export const ENUMS = {
  BrandStatus: {
    values: ['BRAND_STATUS_ACTIVE', 'BRAND_STATUS_INACTIVE'],
    labelKey: (value) => `enums.brandStatus.${value}`,
  },
  PartnerStatus: {
    values: ['PARTNER_STATUS_ACTIVE', 'PARTNER_STATUS_INACTIVE'],
    labelKey: (value) => `enums.partnerStatus.${value}`,
  },
  CategoryStatus: {
    values: ['CATEGORY_STATUS_ACTIVE', 'CATEGORY_STATUS_INACTIVE'],
    labelKey: (value) => `enums.categoryStatus.${value}`,
  },
}

export function getEnumOptions(enumName) {
  const def = ENUMS[enumName]
  if (!def) throw new Error(`Unknown enum: ${enumName}`)
  return def.values
}
