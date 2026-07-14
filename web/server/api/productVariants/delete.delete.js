import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'product_variant.proto',
  service: 'crm.grpc.product_variant.v1.ProductVariantsService',
  entityKey: 'variant',
})

export default defineEventHandler((event) => handlers.remove(event))
