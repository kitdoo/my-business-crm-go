import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'brand.proto',
  service: 'crm.grpc.brand.v1.BrandsService',
  entityKey: 'brand',
})

export default defineEventHandler((event) => handlers.create(event))
