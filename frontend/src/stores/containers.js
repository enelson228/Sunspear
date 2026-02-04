import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useContainersStore = defineStore('containers', () => {
  const containers = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchContainers(all = true) {
    loading.value = true
    error.value = null
    try {
      const response = await api.get(`/containers?all=${all}`)
      containers.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getContainer(id) {
    try {
      const response = await api.get(`/containers/${id}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function startContainer(id) {
    try {
      await api.post(`/containers/${id}/start`)
      await fetchContainers()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function stopContainer(id) {
    try {
      await api.post(`/containers/${id}/stop`)
      await fetchContainers()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function restartContainer(id) {
    try {
      await api.post(`/containers/${id}/restart`)
      await fetchContainers()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function removeContainer(id, force = false) {
    try {
      await api.delete(`/containers/${id}/remove?force=${force}`)
      await fetchContainers()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function getLogs(id, tail = 100) {
    try {
      const response = await api.get(`/containers/${id}/logs?tail=${tail}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    containers,
    loading,
    error,
    fetchContainers,
    getContainer,
    startContainer,
    stopContainer,
    restartContainer,
    removeContainer,
    getLogs
  }
})
