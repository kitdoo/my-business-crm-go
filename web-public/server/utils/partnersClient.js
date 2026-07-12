import { getServiceClient, grpcCall } from './grpcClient.js'

// Server-only, anonymous read of the public partner list (TZ item 13).
// PartnersService.ListPublic is the one PartnersService method exempt from
// auth on the backend and it already forces status=ACTIVE and strips
// commissionPercentage/note server-side — see internal/transports/grpc/
// handlers/partner/handler.go ListPublic — so nothing further needs to be
// filtered here, unlike catalogClient.js which must force the filter itself.
export async function listPublicPartners() {
  const client = getServiceClient('partner.proto', 'crm.grpc.partner.v1.PartnersService')
  const response = await grpcCall(client, 'ListPublic', {})
  return response.items || []
}
