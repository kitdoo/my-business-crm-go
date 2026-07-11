import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'product.proto',
  service: 'crm.grpc.product.v1.ProductsService',
  entityKey: 'product',
})

export default defineEventHandler((event) => handlers.update(event))
