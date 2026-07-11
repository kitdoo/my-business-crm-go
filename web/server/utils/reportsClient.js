import { getServiceClient } from './grpcClient'

// Shared by server/api/reports/*.js — Reports isn't a generic
// EntityRegistry entity (six independent read-only methods, one per
// dashboard widget, TD §8.4), but the routes still all need the same
// service client.
export function getReportsClient() {
  return getServiceClient('report.proto', 'crm.grpc.report.v1.ReportsService')
}
