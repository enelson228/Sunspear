import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useImagesStore = defineStore('images', () => {
  const images = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchImages() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/images')
      images.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function pullImage(imageName) {
    try {
      const response = await api.post('/images/pull', { image: imageName })
      await fetchImages()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function removeImage(id, force = false) {
    try {
      await api.delete(`/images/${id}/remove?force=${force}`)
      await fetchImages()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function searchImages(term) {
    try {
      const response = await api.get(`/images/search?term=${term}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    images,
    loading,
    error,
    fetchImages,
    pullImage,
    removeImage,
    searchImages
  }
})
