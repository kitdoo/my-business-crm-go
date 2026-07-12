import { getServiceClient, grpcCall } from './grpcClient.js'
import { clientKeyHeader } from './clientKey.js'

// Server-only. NotificationsService is a generic outbound-message sender
// on the backend — the message body is opaque, composed entirely by the
// caller (contact/dealer form routes build the text themselves). Send is
// the only NotificationsService method exempt from the backend's user
// bearer-token auth, but it still requires an approved "x-client-key"
// header — see internal/transports/grpc/interceptors/clientkey and
// CRMConfig.NotificationClients.

/**
 * @param {{ subject?: string, message: string, email: string }} msg
 */
export async function sendNotification(msg) {
  const client = getServiceClient('notification.proto', 'crm.grpc.notification.v1.NotificationsService')
  await grpcCall(client, 'Send', msg, { headers: clientKeyHeader() })
}
