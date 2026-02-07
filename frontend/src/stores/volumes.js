import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useVolumesStore = defineStore('volumes', () => {
  const volumes = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchVolumes() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/volumes')
      volumes.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function inspectVolume(name) {
    try {
      const response = await api.get(`/volumes/${name}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function createVolume(config) {
    try {
      const response = await api.post('/volumes', config)
      await fetchVolumes()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function removeVolume(name, force = false) {
    try {
      await api.delete(`/volumes/${name}?force=${force}`)
      await fetchVolumes()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function pruneVolumes() {
    try {
      const response = await api.post('/volumes/prune')
      await fetchVolumes()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    volumes,
    loading,
    error,
    fetchVolumes,
    inspectVolume,
    createVolume,
    removeVolume,
    pruneVolumes
  }
})
