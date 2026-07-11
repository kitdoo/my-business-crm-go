import { storeToRefs } from 'pinia'
import { useAuthStore } from '~/stores/auth'

export function useAuth() {
  const store = useAuthStore()
  const { user, initialized } = storeToRefs(store)
  return {
    user,
    initialized,
    fetchMe: store.fetchMe,
    login: store.login,
    logout: store.logout,
  }
}
