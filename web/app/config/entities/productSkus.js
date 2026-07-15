// Single source of truth for the ProductSKU entity — the purchasable unit
// of a ProductVariant: sku, attributes (characteristics that affect price
// or availability — size, thickness, packaging, …), price (ProductPrice)
// and stock (Inventory) all live here, keyed by skuId, not variantId.
// Images/appearance stay on ProductVariant (see productVariants.js). Never
// shown in the main nav (no `group`) and has no standalone page — always
// rendered inline inside a Variant's expanded block (see
// ProductSkusPanel.vue), which supplies the variantId directly as a prop
// instead of a route param. Price is edited right alongside sku/attributes
// in one compact form (see ProductSkuGeneralForm.vue); stock shows as a
// plain read-only total there and in ProductSkusPanel.vue's collapsed row
// (see utils/inventoryStock.js) — no embedded movements table, that lives
// on the dedicated /inventory-movements page.
export default {
  key: 'productSkus',
  label: 'entities.productSkus.label',
  route: '/product-skus',
  permissions: {
    read: 'productskus:read',
    create: 'productskus:create',
    update: 'productskus:update',
    delete: 'productskus:delete',
  },

  list: {
    columns: [
      { key: 'sku', label: 'fields.sku' },
      { key: 'variantId', label: 'fields.variant', component: 'RelationLabel', relation: 'productVariants' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'productSku' },
    ],
    filters: [{ key: 'statuses', type: 'select', label: 'fields.status', optionsFrom: 'enum:ProductSkuStatus' }],
    defaultFilter: {},
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_SKU', label: 'sort.sku' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  // No status field: ProductSkuGeneralForm.vue flips a new SKU to Active
  // right after Create instead of exposing a draft/inactive toggle (see
  // that component's comment) — Delete is the only other lifecycle action
  // available through the UI.
  form: {
    fields: [
      { key: 'variantId', type: 'relation', relation: 'productVariants', label: 'fields.variant', required: true, createOnly: true },
      { key: 'sku', type: 'text', label: 'fields.sku', required: true, maxLength: 64, immutableOnEdit: true },
      { key: 'attributes', type: 'attributeDetails', label: 'fields.attributes' },
    ],
  },
}
