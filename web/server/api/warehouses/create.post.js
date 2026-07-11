import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'warehouse.proto',
  service: 'crm.grpc.warehouse.v1.WarehousesService',
  entityKey: 'warehouse',
})

export default defineEventHandler((event) => handlers.create(event))
