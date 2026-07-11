// Single source of truth for the Product entity (TD §9.1, §10, §12.1/§12.6)
// — the most complex standard entity: brandId (single relation),
// categoryIds (multi-relation), details (key -> LocalizedString editor),
// imageIds (upload/reorder/remove), sku immutable after creation.
export default {
  key: 'products',
  label: 'entities.products.label',
  icon: 'i-lucide-box',
  route: '/products',
  // Has a tabbed full-page view (General/Price — TD §12.1) beyond the
  // generic form: EntityForm's drawer shows a link to it instead of
  // trying to cram tabs into the drawer.
  detailPage: true,
  permissions: {
    read: 'products:read',
    create: 'products:create',
    update: 'products:update',
    delete: 'products:delete',
  },

  list: {
    columns: [
      { key: 'sku', label: 'fields.sku' },
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'brandId', label: 'fields.brand', component: 'RelationLabel', relation: 'brands' },
      { key: 'status', label: 'fields.status', component: 'StatusBadge', statusMap: 'product' },
      { key: 'createdAt', label: 'fields.createdAt', component: 'DateLabel' },
    ],
    filters: [
      { key: 'statuses', type: 'multiselect', label: 'fields.status', optionsFrom: 'enum:ProductStatus' },
      { key: 'brandIds', type: 'multiselect', label: 'fields.brand', optionsFrom: 'relation:brands' },
      { key: 'categoryIds', type: 'multiselect', label: 'fields.categories', optionsFrom: 'relation:categories' },
      { key: 'createdAt', type: 'periodFilter', label: 'fields.createdAt' },
    ],
    sort: [
      { field: 'FIELD_CREATED_AT', label: 'sort.createdAt' },
      { field: 'FIELD_NAME', label: 'sort.name' },
    ],
    defaultSort: { field: 'FIELD_CREATED_AT', direction: 'SORT_DIRECTION_DESC' },
  },

  form: {
    fields: [
      { key: 'sku', type: 'text', label: 'fields.sku', required: true, maxLength: 64, immutableOnEdit: true },
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'brandId', type: 'relation', relation: 'brands', label: 'fields.brand', required: true },
      { key: 'categoryIds', type: 'relationMulti', relation: 'categories', label: 'fields.categories' },
      { key: 'details', type: 'keyValueLocalized', label: 'fields.details' },
      { key: 'imageIds', type: 'images', label: 'fields.images' },
      { key: 'status', type: 'enum', enum: 'ProductStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
