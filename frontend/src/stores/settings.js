import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useSettingsStore = defineStore('settings', () => {
  const users = ref([])
  const settings = ref({})
  const currentUser = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function fetchCurrentUser() {
    try {
      const response = await api.get('/auth/me')
      currentUser.value = response.data
    } catch (err) {
      console.error('Failed to fetch current user:', err)
    }
  }

  async function fetchUsers() {
    loading.value = true
    try {
      const response = await api.get('/users')
      users.value = response.data
    } catch (err) {
      error.value = 'Failed to fetch users'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createUser(username, password) {
    const response = await api.post('/users', { username, password })
    await fetchUsers()
    return response.data
  }

  async function deleteUser(id) {
    await api.delete(`/users/${id}`)
    await fetchUsers()
  }

  async function changePassword(userId, currentPassword, newPassword) {
    await api.put(`/users/${userId}/password`, {
      current_password: currentPassword,
      new_password: newPassword
    })
  }

  async function fetchSettings() {
    try {
      const response = await api.get('/settings')
      settings.value = response.data
    } catch (err) {
      console.error('Failed to fetch settings:', err)
    }
  }

  async function updateSettings(newSettings) {
    await api.put('/settings', newSettings)
    settings.value = { ...settings.value, ...newSettings }
  }

  return {
    users, settings, currentUser, loading, error,
    fetchCurrentUser, fetchUsers, createUser, deleteUser,
    changePassword, fetchSettings, updateSettings
  }
})
