import { getServiceClient, grpcCall } from '~~/server/utils/grpcClient'
import { mapGrpcError } from '~~/server/utils/mapGrpcError'
import { requireSession } from '~~/server/utils/session'

/**
 * One factory generates the 5 standard CRUD server routes for any entity
 * (TD §9.2) — server/api/[entity]/*.js files stay one-liners that just
 * call the matching function returned here. Wire contract expected from
 * app/composables/useEntityApi.js:
 *
 *   POST   /api/<entity>/list    { sort?, pagination?, filter? } -> { items, total, nextCursor }
 *   POST   /api/<entity>/get     { id } -> <entity>
 *   POST   /api/<entity>/create  { fields } -> <entity>
 *   PATCH  /api/<entity>/update  { id, fields, mask, etag } -> <entity>
 *   DELETE /api/<entity>/delete  { id, etag } -> { ok: true }
 *
 * @param {{ protoFile: string, service: string, entityKey: string }} config
 *   entityKey is the response wrapper field, e.g. 'brand' for
 *   BrandGetResponse.brand / BrandCreateResponse.brand / BrandUpdateResponse.brand.
 */
export function defineEntityHandler({ protoFile, service, entityKey }) {
  function getClient() {
    return getServiceClient(protoFile, service)
  }

  async function callMethod(event, methodName, buildRequest, unwrap) {
    const session = requireSession(event)
    const body = (await readBody(event).catch(() => ({}))) || {}
    const request = buildRequest(body)
    const client = getClient()
    try {
      const response = await grpcCall(client, methodName, request, { token: session.token })
      return unwrap ? unwrap(response) : response
    } catch (err) {
      throw mapGrpcError(err)
    }
  }

  return {
    list: (event) =>
      callMethod(
        event,
        'List',
        (body) => ({
          sort: body.sort,
          pagination: body.pagination,
          filter: body.filter,
          options: { includeTotalCount: true },
        }),
        (res) => ({ items: res.items || [], total: res.total, nextCursor: res.nextCursor }),
      ),

    get: (event) =>
      callMethod(
        event,
        'Get',
        (body) => ({ id: body.id }),
        (res) => res[entityKey],
      ),

    create: (event) =>
      callMethod(
        event,
        'Create',
        (body) => ({ ...body.fields }),
        (res) => res[entityKey],
      ),

    update: (event) =>
      callMethod(
        event,
        'Update',
        (body) => ({
          id: body.id,
          ...body.fields,
          options: { updateMask: { paths: body.mask || [] }, etag: body.etag },
        }),
        (res) => res[entityKey],
      ),

    remove: (event) =>
      callMethod(
        event,
        'Delete',
        (body) => ({ id: body.id, options: { etag: body.etag } }),
        () => ({ ok: true }),
      ),
  }
}
