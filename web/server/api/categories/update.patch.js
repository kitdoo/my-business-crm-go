import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'category.proto',
  service: 'crm.grpc.category.v1.CategoriesService',
  entityKey: 'category',
})

export default defineEventHandler((event) => handlers.update(event))
