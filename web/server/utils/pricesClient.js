import { getServiceClient } from './grpcClient'

// Shared by server/api/prices/*.js — ProductPrice isn't a generic
// EntityRegistry entity (Get is keyed by variantId, not id), so its
// routes are hand-written rather than generated via defineEntityHandler,
// but they still all need the same service client.
export function getPricesClient() {
  return getServiceClient('price.proto', 'crm.grpc.price.v1.PricesService')
}
