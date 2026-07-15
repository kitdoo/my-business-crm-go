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
  UserStatus: {
    values: ['USER_STATUS_ACTIVE', 'USER_STATUS_INACTIVE'],
    labelKey: (value) => `enums.userStatus.${value}`,
  },
  UserRole: {
    values: ['USER_ROLE_ADMIN', 'USER_ROLE_EMPLOYEE', 'USER_ROLE_GUEST'],
    labelKey: (value) => `enums.userRole.${value}`,
  },
  ProductStatus: {
    values: ['PRODUCT_STATUS_DRAFT', 'PRODUCT_STATUS_ACTIVE', 'PRODUCT_STATUS_INACTIVE'],
    labelKey: (value) => `enums.productStatus.${value}`,
  },
  ProductVariantStatus: {
    values: ['PRODUCT_VARIANT_STATUS_DRAFT', 'PRODUCT_VARIANT_STATUS_ACTIVE', 'PRODUCT_VARIANT_STATUS_INACTIVE'],
    labelKey: (value) => `enums.productVariantStatus.${value}`,
  },
  ProductSkuStatus: {
    values: ['PRODUCT_SKU_STATUS_DRAFT', 'PRODUCT_SKU_STATUS_ACTIVE', 'PRODUCT_SKU_STATUS_INACTIVE'],
    labelKey: (value) => `enums.productSkuStatus.${value}`,
  },
  WarehouseStatus: {
    values: ['WAREHOUSE_STATUS_ACTIVE', 'WAREHOUSE_STATUS_INACTIVE'],
    labelKey: (value) => `enums.warehouseStatus.${value}`,
  },
  // MOVEMENT_TYPE_SALE is deliberately excluded — it's an internal type
  // created only by SalesService.Create, never picked by hand (TD §12.4).
  MovementType: {
    values: ['MOVEMENT_TYPE_RECEIPT', 'MOVEMENT_TYPE_WRITE_OFF', 'MOVEMENT_TYPE_ADJUSTMENT', 'MOVEMENT_TYPE_TRANSFER'],
    labelKey: (value) => `enums.movementType.${value}`,
  },
  SaleStatus: {
    values: [
      'SALE_STATUS_DRAFT',
      'SALE_STATUS_PAID',
      'SALE_STATUS_SHIPPED',
      'SALE_STATUS_COMPLETED',
      'SALE_STATUS_CANCELLED',
      'SALE_STATUS_REFUNDED',
    ],
    labelKey: (value) => `enums.saleStatus.${value}`,
  },
}

export function getEnumOptions(enumName) {
  const def = ENUMS[enumName]
  if (!def) throw new Error(`Unknown enum: ${enumName}`)
  return def.values
}
