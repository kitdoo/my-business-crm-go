import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'client.proto',
  service: 'crm.grpc.client.v1.ClientsService',
  entityKey: 'client',
})

export default defineEventHandler((event) => handlers.remove(event))
