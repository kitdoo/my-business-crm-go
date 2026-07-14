// Single source of truth for the ProductVariant entity — the purchasable
// unit of a Product: sku, attributes (differentiating characteristics —
// color, size, …), imageIds, price (ProductPrice) and stock (Inventory)
// all live here, keyed by variantId, not productId. Never shown in the
// main nav (no `group`) — always reached from within a Product's page
// (see ProductVariantsTab.vue), which supplies the `fixedFilter`/
// `createDefaults` productId. Has its own tabbed full-page view (General/
// Price/Movements), same rationale and shape as Product's own detail page
// used to have before sku/price/stock moved here.
export default {
  key: 'productVariants',
  label: 'entities.productVariants.label',
  route: '/product-variants',
  detailPage: true,
  permissions: {
    read: 'productvariants:read',
    create: 'productvariants:create',
    update: 'productvariants:update',
    delete: 'productvariants:delete',
  },

  list: {
    columns: [
      { key: 'imageIds', label: 'fields.images', component: 'ImageThumbnail' },
      { key: 'sku', label: 'fields.sku' },
      { key: 'productId', label: 'fields.product', component: 'RelationLabel', relation: 'products' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'productVariant' },
    ],
    filters: [{ key: 'statuses', type: 'select', label: 'fields.status', optionsFrom: 'enum:ProductVariantStatus' }],
    defaultFilter: {},
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_SKU', label: 'sort.sku' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  form: {
    fields: [
      { key: 'productId', type: 'relation', relation: 'products', label: 'fields.product', required: true, createOnly: true },
      { key: 'sku', type: 'text', label: 'fields.sku', required: true, maxLength: 64, immutableOnEdit: true },
      { key: 'attributes', type: 'attributeDetails', label: 'fields.attributes' },
      { key: 'imageIds', type: 'images', label: 'fields.images' },
      { key: 'status', type: 'enum', enum: 'ProductVariantStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
