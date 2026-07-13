// Single source of truth for the Product entity (TD §9.1, §10, §12.1/§12.6)
// — the most complex standard entity: brandId (single relation),
// categoryIds (multi-relation), details (backend-defined characteristics
// catalog, see AttributeDetailsInput.vue), imageIds (upload/reorder/remove),
// sku immutable after creation. Create/edit use the bespoke
// ProductGeneralForm.vue, not the generic <EntityForm>, for its grouped
// block layout — form.fields below still backs useEntityForm's
// load/save/etag/mask logic.
export default {
  key: 'products',
  label: 'entities.products.label',
  icon: 'i-lucide-box',
  route: '/products',
  group: 'catalog',
  // Has a tabbed full-page view (General/Price — TD §12.1) beyond the
  // generic form: EntityForm's drawer shows a link to it instead of
  // trying to cram tabs into the drawer.
  detailPage: true,
  // View drawer shows read-only latest price + stock-by-warehouse under
  // the field list — see EntityViewDrawer.vue.
  view: {
    productSummary: true,
  },
  permissions: {
    read: 'products:read',
    create: 'products:create',
    update: 'products:update',
    delete: 'products:delete',
  },

  list: {
    columns: [
      { key: 'imageIds', label: 'fields.images', component: 'ImageThumbnail' },
      { key: 'sku', label: 'fields.sku' },
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'brandId', label: 'fields.brand', component: 'RelationLabel', relation: 'brands' },
      { key: 'categoryIds', label: 'fields.categories', component: 'RelationListLabel', relation: 'categories' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'product' },
    ],
    filters: [
      // Single-select at the top of the table, not a multiselect — per
      // the request, picking one category to narrow the list, not
      // filtering by a combination.
      { key: 'categoryIds', type: 'select', label: 'fields.categories', optionsFrom: 'relation:categories' },
      { key: 'statuses', type: 'select', label: 'fields.status', optionsFrom: 'enum:ProductStatus' },
      { key: 'brandIds', type: 'multiselect', label: 'fields.brand', optionsFrom: 'relation:brands' },
    ],
    defaultFilter: { statuses: 'PRODUCT_STATUS_ACTIVE' },
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_NAME', label: 'sort.name' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  form: {
    // Order per the request: brand, categories, sku, name, description,
    // details, images (status stays editOnly, unaffected).
    fields: [
      { key: 'brandId', type: 'relation', relation: 'brands', label: 'fields.brand', required: true },
      { key: 'categoryIds', type: 'relationMulti', relation: 'categories', label: 'fields.categories' },
      { key: 'sku', type: 'text', label: 'fields.sku', required: true, maxLength: 64, immutableOnEdit: true },
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'details', type: 'attributeDetails', label: 'fields.details' },
      { key: 'imageIds', type: 'images', label: 'fields.images' },
      { key: 'status', type: 'enum', enum: 'ProductStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
