import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'user.proto',
  service: 'crm.grpc.user.v1.UsersService',
  entityKey: 'user',
})

export default defineEventHandler((event) => handlers.list(event))
