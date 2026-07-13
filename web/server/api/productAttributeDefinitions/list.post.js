import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

// Read-only characteristics catalog (TZ) — no admin CRUD UI, so only the
// list route exists, same as web/server/api/inventory/list.post.js.
const handlers = defineEntityHandler({
  protoFile: 'product_attribute_definition.proto',
  service: 'crm.grpc.product_attribute_definition.v1.ProductAttributeDefinitionsService',
  entityKey: 'productAttributeDefinition',
})

export default defineEventHandler((event) => handlers.list(event))
