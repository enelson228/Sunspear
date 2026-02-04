import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useSystemStore = defineStore('system', () => {
  const metrics = ref(null)
  const systemInfo = ref(null)
  const loading = ref(false)
  const error = ref(null)

  let pollingInterval = null

  async function fetchMetrics() {
    try {
      const response = await api.get('/system/metrics')
      metrics.value = response.data
    } catch (err) {
      error.value = err.message
    }
  }

  async function fetchSystemInfo() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/system/info')
      systemInfo.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  function startPolling(interval = 5000) {
    stopPolling()
    fetchMetrics()
    pollingInterval = setInterval(fetchMetrics, interval)
  }

  function stopPolling() {
    if (pollingInterval) {
      clearInterval(pollingInterval)
      pollingInterval = null
    }
  }

  return {
    metrics,
    systemInfo,
    loading,
    error,
    fetchMetrics,
    fetchSystemInfo,
    startPolling,
    stopPolling
  }
})
