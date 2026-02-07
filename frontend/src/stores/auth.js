import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/composables/useDockerAPI'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const isAuthenticated = computed(() => !!token.value)

  async function login(username, password) {
    try {
      const response = await api.post('/auth/login', { username, password })
      token.value = response.data.token
      localStorage.setItem('token', token.value)
      return true
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    }
  }

  async function setup(username, password) {
    try {
      await api.post('/auth/setup', { username, password })
      return await login(username, password)
    } catch (error) {
      console.error('Setup failed:', error)
      throw error
    }
  }

  async function checkSetupRequired() {
    try {
      const response = await api.get('/auth/setup/status')
      return !!response.data?.needs_setup
    } catch (error) {
      console.error('Setup status check failed:', error)
      return false
    }
  }

  function logout() {
    token.value = null
    localStorage.removeItem('token')
  }

  function getToken() {
    return token.value
  }

  return {
    token,
    isAuthenticated,
    login,
    setup,
    checkSetupRequired,
    logout,
    getToken
  }
})
