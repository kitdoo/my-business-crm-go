import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'product_sku.proto',
  service: 'crm.grpc.product_sku.v1.ProductSKUsService',
  entityKey: 'sku',
})

export default defineEventHandler((event) => handlers.update(event))
