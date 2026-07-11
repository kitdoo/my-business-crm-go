// Single place server-route errors get turned into UI reactions (TD §4.4/
// §9.4). No component should write its own try/catch with custom error
// text — always route through handle() so messages stay consistent.
export function useApiErrorHandler() {
  const toast = useToast()
  const { t } = useI18n()
  const route = useRoute()

  /**
   * @param {unknown} error a FetchError thrown by $fetch against our
   *   server routes — error.data.error is { code, message, field? }.
   * @param {{ onFieldError?: (field: string, message: string) => void,
   *           onEtagConflict?: () => void }} [handlers]
   */
  function handle(error, handlers = {}) {
    const apiError = error?.data?.error
    if (!apiError) {
      toast.add({ title: t('errors.generic'), color: 'error' })
      console.error(error)
      return
    }

    const { code, message, field } = apiError

    switch (code) {
      case 'NOT_FOUND':
        toast.add({ title: t('errors.notFound'), color: 'error' })
        break
      case 'ALREADY_EXISTS':
        if (field && handlers.onFieldError) {
          handlers.onFieldError(field, t('errors.alreadyExists'))
        } else {
          toast.add({ title: t('errors.alreadyExists'), color: 'error' })
        }
        break
      case 'ABORTED':
        if (handlers.onEtagConflict) {
          handlers.onEtagConflict()
        } else {
          toast.add({ title: t('errors.staleEntity'), color: 'error' })
        }
        break
      case 'FAILED_PRECONDITION':
        toast.add({ title: message, color: 'error' })
        break
      case 'INVALID_ARGUMENT':
        if (field && handlers.onFieldError) {
          handlers.onFieldError(field, message)
        } else {
          toast.add({ title: message, color: 'error' })
        }
        break
      case 'UNAUTHENTICATED':
        navigateTo({ path: '/login', query: { redirect: route.fullPath } })
        break
      case 'PERMISSION_DENIED':
        toast.add({ title: t('errors.permissionDenied'), color: 'error' })
        break
      case 'UNIMPLEMENTED':
        toast.add({ title: t('errors.unimplemented'), color: 'warning' })
        break
      default:
        toast.add({ title: t('errors.generic'), color: 'error' })
        console.error(apiError)
    }
  }

  return { handle }
}
