import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useNetworksStore = defineStore('networks', () => {
  const networks = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchNetworks() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/networks')
      networks.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function inspectNetwork(id) {
    try {
      const response = await api.get(`/networks/${id}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function createNetwork(config) {
    try {
      const response = await api.post('/networks', config)
      await fetchNetworks()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function removeNetwork(id) {
    try {
      await api.delete(`/networks/${id}`)
      await fetchNetworks()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function connectContainer(networkId, containerId) {
    try {
      const response = await api.post(`/networks/${networkId}/connect`, { container_id: containerId })
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function disconnectContainer(networkId, containerId) {
    try {
      const response = await api.post(`/networks/${networkId}/disconnect`, { container_id: containerId })
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function pruneNetworks() {
    try {
      const response = await api.post('/networks/prune')
      await fetchNetworks()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    networks,
    loading,
    error,
    fetchNetworks,
    inspectNetwork,
    createNetwork,
    removeNetwork,
    connectContainer,
    disconnectContainer,
    pruneNetworks
  }
})
