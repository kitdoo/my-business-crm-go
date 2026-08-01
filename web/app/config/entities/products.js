// Single source of truth for the Product entity (TD §9.1, §10, §12.1/§12.6)
// — a catalog card grouping one or more ProductVariant: brandId (single
// relation), categoryIds (multi-relation), details (backend-defined
// characteristics catalog, shared across every variant — see
// AttributeDetailsInput.vue). sku/imageIds/price/stock all live on
// ProductVariant instead (see config/entities/productVariants.js) — a
// product on its own is not purchasable. Create/edit use the bespoke
// ProductGeneralForm.vue, not the generic <EntityForm>, for its grouped
// block layout — form.fields below still backs useEntityForm's
// load/save/etag/mask logic.
export default {
  key: 'products',
  label: 'entities.products.label',
  icon: 'i-lucide-box',
  route: '/products',
  group: 'catalog',
  // Has a tabbed full-page view (General/Variants — TD §12.1) beyond the
  // generic form: EntityForm's drawer shows a link to it instead of
  // trying to cram tabs into the drawer.
  detailPage: true,
  permissions: {
    read: 'products:read',
    create: 'products:create',
    update: 'products:update',
    delete: 'products:delete',
  },

  // Read-only variants/SKUs/stock underneath the drawer's own fields (see
  // EntityViewDrawer.vue / ProductVariantsReadOnly.vue) — editing any of it
  // still only happens on the full /products/:id page.
  view: {
    variantsSummary: true,
  },

  list: {
    // Lets useReferenceCacheStore batch RelationLabel lookups (e.g.
    // productVariants' "product" column) into one filter.ids List call
    // instead of one Get per row — backend already supports it
    // (ProductsListRequest.Filter.ids).
    idsFilterKey: 'ids',
    columns: [
      { key: 'name', label: 'fields.name', component: 'LocalizedText' },
      { key: 'brandId', label: 'fields.brand', component: 'RelationLabel', relation: 'brands' },
      { key: 'categoryIds', label: 'fields.categories', component: 'RelationListLabel', relation: 'categories' },
      { key: 'priceUnit', label: 'fields.priceUnit', component: 'StatusBadge', statusMap: 'priceUnit' },
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
    // Order per the request: brand, categories, name, description,
    // details (status stays editOnly, unaffected).
    fields: [
      { key: 'brandId', type: 'relation', relation: 'brands', label: 'fields.brand', required: true, searchable: true },
      { key: 'categoryIds', type: 'relationMulti', relation: 'categories', label: 'fields.categories' },
      { key: 'name', type: 'localizedString', label: 'fields.name', required: true },
      { key: 'description', type: 'localizedString', label: 'fields.description' },
      { key: 'priceUnit', type: 'enum', enum: 'PriceUnit', label: 'fields.priceUnit', required: true },
      { key: 'details', type: 'attributeDetails', label: 'fields.details' },
      { key: 'status', type: 'enum', enum: 'ProductStatus', label: 'fields.status', editOnly: true },
    ],
  },
}
