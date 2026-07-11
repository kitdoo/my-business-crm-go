import { defineEntityHandler } from '~~/server/utils/defineEntityHandler'

const handlers = defineEntityHandler({
  protoFile: 'partner.proto',
  service: 'crm.grpc.partner.v1.PartnersService',
  entityKey: 'partner',
})

export default defineEventHandler((event) => handlers.create(event))
