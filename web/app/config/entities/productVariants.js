// Single source of truth for the ProductVariant entity — the visual
// identity of a Product: attributes that change appearance (color,
// texture, pattern, …) and imageIds. Carries no sku/price/stock — those
// live on ProductSKU (see productSkus.js), one or more of which belong to
// each variant. Never shown in the main nav (no `group`) and has no
// standalone page — always rendered inline on the Product page (see
// ProductVariantsPanel.vue / ProductSkusPanel.vue), which supplies the
// productId/variantId directly as props instead of a route param.
export default {
  key: 'productVariants',
  label: 'entities.productVariants.label',
  route: '/product-variants',
  permissions: {
    read: 'productvariants:read',
    create: 'productvariants:create',
    update: 'productvariants:update',
    delete: 'productvariants:delete',
  },

  list: {
    columns: [
      { key: 'imageIds', label: 'fields.images', component: 'ImageThumbnail' },
      { key: 'productId', label: 'fields.product', component: 'RelationLabel', relation: 'products' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'productVariant' },
    ],
    filters: [{ key: 'statuses', type: 'select', label: 'fields.status', optionsFrom: 'enum:ProductVariantStatus' }],
    defaultFilter: {},
    sort: [{ field: 'FIELD_CREATED_AT', label: 'sort.createdAt' }],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  // No status field: ProductVariantGeneralForm.vue flips a new variant to
  // Active right after Create instead of exposing a draft/inactive toggle
  // (see that component's comment) — Delete is the only other lifecycle
  // action available through the UI.
  form: {
    fields: [
      {
        key: 'productId',
        type: 'relation',
        relation: 'products',
        label: 'fields.product',
        required: true,
        createOnly: true,
        searchable: true,
      },
      { key: 'attributes', type: 'attributeDetails', label: 'fields.attributes' },
      { key: 'imageIds', type: 'images', label: 'fields.images' },
    ],
  },
}
