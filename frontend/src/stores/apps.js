import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useAppsStore = defineStore('apps', () => {
  const apps = ref([])
  const installedApps = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchApps() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/apps')
      apps.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getApp(id) {
    try {
      const response = await api.get(`/apps/${id}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function installApp(id, config) {
    try {
      const response = await api.post(`/apps/${id}/install`, config)
      await fetchInstalledApps()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function fetchInstalledApps() {
    try {
      const response = await api.get('/apps/installed')
      installedApps.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function uninstallApp(installedId) {
    try {
      await api.post(`/apps/installed/${installedId}/uninstall`)
      await fetchInstalledApps()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    apps,
    installedApps,
    loading,
    error,
    fetchApps,
    getApp,
    installApp,
    fetchInstalledApps,
    uninstallApp
  }
})
