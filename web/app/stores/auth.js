import { defineStore } from 'pinia'

// Holds only the public User fields the BFF returns — the gRPC token
// never reaches the browser (TD §2/§4.1).
export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    initialized: false,
  }),
  actions: {
    async fetchMe() {
      // useRequestFetch (not plain $fetch) forwards the incoming request's
      // cookies during SSR — without it, the first server-rendered
      // navigation never sees the session cookie and the auth middleware
      // wrongly redirects to /login even when a valid session exists.
      const requestFetch = useRequestFetch()
      const { user } = await requestFetch('/api/auth/me')
      this.user = user
      this.initialized = true
      return user
    },
    async login(login, password) {
      const { user } = await $fetch('/api/auth/login', {
        method: 'POST',
        body: { login, password },
      })
      this.user = user
      this.initialized = true
      return user
    },
    async logout() {
      await $fetch('/api/auth/logout', { method: 'POST' })
      this.user = null
    },
  },
})
