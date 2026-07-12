// Server-only. The static API key this frontend is registered under on
// the backend (CRMConfig.NotificationClients) — see notificationClient.js.
export function clientKeyHeader() {
  const config = useRuntimeConfig()
  const key = config.grpc.clientKey
  if (!key) {
    throw new Error('runtimeConfig.grpc.clientKey is not set (env NUXT_GRPC_CLIENT_KEY)')
  }
  return { 'x-client-key': key }
}
