import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/composables/useDockerAPI'

export const useComposeStore = defineStore('compose', () => {
  const projects = ref([])
  const templates = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchProjects() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/compose/projects')
      projects.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function getProject(id) {
    try {
      const response = await api.get(`/compose/projects/${id}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function deployProject(config) {
    try {
      const response = await api.post('/compose/projects', config)
      await fetchProjects()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function validateYAML(yaml) {
    try {
      const response = await api.post('/compose/validate', { yaml })
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function startProject(id) {
    try {
      const response = await api.post(`/compose/projects/${id}/start`)
      await fetchProjects()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function stopProject(id) {
    try {
      const response = await api.post(`/compose/projects/${id}/stop`)
      await fetchProjects()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function restartProject(id) {
    try {
      const response = await api.post(`/compose/projects/${id}/restart`)
      await fetchProjects()
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function deleteProject(id) {
    try {
      await api.delete(`/compose/projects/${id}`)
      await fetchProjects()
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function fetchTemplates() {
    try {
      const response = await api.get('/compose/templates')
      templates.value = response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  async function getTemplate(name) {
    try {
      const response = await api.get(`/compose/templates/${name}`)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  return {
    projects,
    templates,
    loading,
    error,
    fetchProjects,
    getProject,
    deployProject,
    validateYAML,
    startProject,
    stopProject,
    restartProject,
    deleteProject,
    fetchTemplates,
    getTemplate
  }
})
