// Global route guard (TD §4.1 step 5): every page except /login requires a
// valid session. Runs on both server (first load) and client (SPA nav).
export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === '/login') return

  const { user, initialized, fetchMe } = useAuth()
  if (!initialized.value) {
    await fetchMe()
  }
  if (!user.value) {
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  }
})
