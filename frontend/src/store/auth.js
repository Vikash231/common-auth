import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('auth_token') || null,
    user: JSON.parse(localStorage.getItem('auth_user') || 'null'),
    loading: false,
    error: null
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    isAdmin: (state) => state.user?.is_admin || false,
    permissions: (state) => state.user?.permissions || []
  },

  actions: {
    setToken(token) {
      this.token = token
      localStorage.setItem('auth_token', token)
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`
    },

    setUser(user) {
      this.user = user
      localStorage.setItem('auth_user', JSON.stringify(user))
    },

    async login(email, password) {
      this.loading = true
      this.error = null
      try {
        const res = await api.post('/api/auth/login', { email, password })
        this.setToken(res.data.token)
        await this.fetchMe()
        return true
      } catch (err) {
        this.error = err.response?.data?.error || 'Login failed'
        return false
      } finally {
        this.loading = false
      }
    },

    async fetchMe() {
      try {
        const res = await api.get('/api/auth/me')
        this.setUser(res.data)
      } catch {
        this.logout()
      }
    },

    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_user')
      delete api.defaults.headers.common['Authorization']
    },

    initFromStorage() {
      if (this.token) {
        api.defaults.headers.common['Authorization'] = `Bearer ${this.token}`
      }
    }
  }
})
