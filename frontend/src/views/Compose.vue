<template>
  <div class="compose-page">
    <main class="compose-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>COMPOSE DEPLOYMENTS</h1>
            <p class="subtitle">Deploy and manage multi-container stacks</p>
          </div>
          <div class="header-actions">
            <Button variant="primary" @click="showDeployModal = true">
              + DEPLOY PROJECT
            </Button>
          </div>
        </div>

        <div class="tabs-section">
          <button
            class="tab-button"
            :class="{ active: activeTab === 'projects' }"
            @click="activeTab = 'projects'"
          >
            PROJECTS
          </button>
          <button
            class="tab-button"
            :class="{ active: activeTab === 'templates' }"
            @click="activeTab = 'templates'; loadTemplates()"
          >
            TEMPLATES
          </button>
        </div>

        <!-- Projects Tab -->
        <div v-if="activeTab === 'projects'" class="tab-content">
          <div v-if="loading && projects.length === 0" class="loading-state">
            <div class="spinner"></div>
            <p>Loading projects...</p>
          </div>

          <div v-else-if="error" class="error-state">
            <p class="error-message">{{ error }}</p>
            <Button variant="primary" @click="fetchProjects">RETRY</Button>
          </div>

          <div v-else-if="projects.length === 0" class="empty-state">
            <h3>NO PROJECTS DEPLOYED</h3>
            <p class="text-secondary">Deploy a project to get started</p>
            <Button variant="primary" @click="showDeployModal = true">
              DEPLOY YOUR FIRST PROJECT
            </Button>
          </div>

          <div v-else class="projects-grid">
            <ProjectCard
              v-for="project in projects"
              :key="project.id"
              :project="project"
              :loading="actionLoading === project.id"
              @start="handleStart"
              @stop="handleStop"
              @restart="handleRestart"
              @details="handleDetails"
              @delete="handleDelete"
            />
          </div>
        </div>

        <!-- Templates Tab -->
        <div v-if="activeTab === 'templates'" class="tab-content">
          <div v-if="loadingTemplates" class="loading-state">
            <div class="spinner"></div>
            <p>Loading templates...</p>
          </div>

          <div v-else-if="templates.length === 0" class="empty-state">
            <h3>NO TEMPLATES AVAILABLE</h3>
            <p class="text-secondary">No compose templates found</p>
          </div>

          <div v-else class="templates-grid">
            <Card
              v-for="template in templates"
              :key="template.name"
              class="template-card accent-bar"
              :show-corners="true"
            >
              <h3 class="template-name">{{ template.name }}</h3>
              <p class="template-description">{{ template.description }}</p>
              <Button
                variant="primary"
                @click="handleUseTemplate(template.name)"
              >
                USE TEMPLATE
              </Button>
            </Card>
          </div>
        </div>
      </div>
    </main>

    <!-- Deploy Project Modal -->
    <Modal v-model="showDeployModal" title="DEPLOY PROJECT" wide>
      <div class="deploy-form">
        <Input
          v-model="deployForm.name"
          label="PROJECT NAME *"
          placeholder="my-project"
          type="text"
        />
        <Input
          v-model="deployForm.description"
          label="DESCRIPTION"
          placeholder="Brief description of this project"
          type="text"
        />
        <ComposeEditor
          v-model="deployForm.yaml"
          label="DOCKER COMPOSE YAML *"
          :validating="validating"
          :validation-result="validationResult"
          @validate="handleValidate"
        />
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeDeployModal">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleDeploy" :loading="deploying">
          {{ deploying ? 'DEPLOYING...' : 'DEPLOY PROJECT' }}
        </Button>
      </template>
    </Modal>

    <!-- Project Details Modal -->
    <Modal v-model="showDetailsModal" title="PROJECT DETAILS">
      <div v-if="loadingDetails" class="loading-state">
        <div class="spinner"></div>
        <p>Loading project details...</p>
      </div>
      <div v-else-if="projectDetails" class="details-content">
        <div class="details-header">
          <div>
            <h3 class="project-title">{{ projectDetails.name }}</h3>
            <p class="project-desc" v-if="projectDetails.description">{{ projectDetails.description }}</p>
          </div>
          <Badge :variant="getStatusVariant(projectDetails.status)">{{ projectDetails.status || 'UNKNOWN' }}</Badge>
        </div>

        <div class="details-section">
          <h4 class="section-title">CONTAINERS ({{ projectDetails.container_ids?.length || 0 }})</h4>
          <div v-if="projectDetails.container_ids?.length" class="container-list">
            <div v-for="containerId in projectDetails.container_ids" :key="containerId" class="container-item font-mono">
              {{ containerId }}
            </div>
          </div>
          <p v-else class="text-secondary">No containers</p>
        </div>

        <div class="details-section">
          <h4 class="section-title">ACTIONS</h4>
          <div class="details-actions">
            <Button
              v-if="projectDetails.status !== 'running'"
              variant="primary"
              @click="handleStart(projectDetails.id)"
              :loading="actionLoading === projectDetails.id"
            >
              START
            </Button>
            <Button
              v-if="projectDetails.status === 'running'"
              variant="secondary"
              @click="handleStop(projectDetails.id)"
              :loading="actionLoading === projectDetails.id"
            >
              STOP
            </Button>
            <Button
              v-if="projectDetails.status === 'running'"
              variant="secondary"
              @click="handleRestart(projectDetails.id)"
              :loading="actionLoading === projectDetails.id"
            >
              RESTART
            </Button>
          </div>
        </div>

        <div class="details-section">
          <h4 class="section-title">COMPOSE YAML</h4>
          <pre class="yaml-display">{{ projectDetails.yaml || 'N/A' }}</pre>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeDetailsModal">
          CLOSE
        </Button>
      </template>
    </Modal>

    <!-- Delete Confirmation Modal -->
    <Modal v-model="showDeleteModal" title="CONFIRM DELETE">
      <p>Are you sure you want to delete this project?</p>
      <p class="text-secondary mt-sm">This will remove all containers, networks, and volumes created by this project. This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showDeleteModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="confirmDelete" :loading="actionLoading">
          {{ actionLoading ? 'DELETING...' : 'DELETE' }}
        </Button>
      </template>
    </Modal>

    <!-- Toast Notification -->
    <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useComposeStore } from '@/stores/compose'
import ProjectCard from '@/components/compose/ProjectCard.vue'
import ComposeEditor from '@/components/compose/ComposeEditor.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const composeStore = useComposeStore()

const activeTab = ref('projects')
const actionLoading = ref(null)
const showDeployModal = ref(false)
const showDetailsModal = ref(false)
const showDeleteModal = ref(false)
const deploying = ref(false)
const validating = ref(false)
const loadingDetails = ref(false)
const loadingTemplates = ref(false)
const projectToDelete = ref(null)
const projectDetails = ref(null)
const validationResult = ref(null)
const deployForm = ref({ name: '', description: '', yaml: '' })
const toast = ref({ show: false, message: '', type: 'success' })

const projects = computed(() => composeStore.projects)
const templates = computed(() => composeStore.templates)
const loading = computed(() => composeStore.loading)
const error = computed(() => composeStore.error)

async function fetchProjects() {
  try {
    await composeStore.fetchProjects()
  } catch (err) {
    showToast('Failed to fetch projects', 'error')
  }
}

async function loadTemplates() {
  if (templates.value.length > 0) return

  loadingTemplates.value = true
  try {
    await composeStore.fetchTemplates()
  } catch (err) {
    showToast('Failed to load templates', 'error')
  } finally {
    loadingTemplates.value = false
  }
}

async function handleValidate() {
  if (!deployForm.value.yaml) {
    validationResult.value = { type: 'error', message: 'YAML is required' }
    return
  }

  validating.value = true
  validationResult.value = null
  try {
    const result = await composeStore.validateYAML(deployForm.value.yaml)
    if (result.valid) {
      const serviceCount = result.service_count || 0
      validationResult.value = {
        type: 'success',
        message: `Valid - ${serviceCount} service${serviceCount !== 1 ? 's' : ''} found`
      }
    } else {
      validationResult.value = {
        type: 'error',
        message: `Invalid: ${result.error || 'Unknown error'}`
      }
    }
  } catch (err) {
    validationResult.value = {
      type: 'error',
      message: 'Validation failed'
    }
  } finally {
    validating.value = false
  }
}

async function handleDeploy() {
  if (!deployForm.value.name || !deployForm.value.yaml) {
    showToast('Project name and YAML are required', 'error')
    return
  }

  deploying.value = true
  try {
    await composeStore.deployProject({
      name: deployForm.value.name,
      description: deployForm.value.description,
      yaml: deployForm.value.yaml
    })
    showToast(`Project ${deployForm.value.name} deployed successfully`, 'success')
    closeDeployModal()
  } catch (err) {
    showToast('Failed to deploy project', 'error')
  } finally {
    deploying.value = false
  }
}

function closeDeployModal() {
  showDeployModal.value = false
  deployForm.value = { name: '', description: '', yaml: '' }
  validationResult.value = null
}

async function handleUseTemplate(templateName) {
  loadingTemplates.value = true
  try {
    const template = await composeStore.getTemplate(templateName)
    deployForm.value = {
      name: '',
      description: template.description || '',
      yaml: template.yaml || ''
    }
    showDeployModal.value = true
  } catch (err) {
    showToast('Failed to load template', 'error')
  } finally {
    loadingTemplates.value = false
  }
}

async function handleStart(id) {
  actionLoading.value = id
  try {
    await composeStore.startProject(id)
    showToast('Project started successfully', 'success')
    if (showDetailsModal.value && projectDetails.value?.id === id) {
      await loadProjectDetails(id)
    }
  } catch (err) {
    showToast('Failed to start project', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleStop(id) {
  actionLoading.value = id
  try {
    await composeStore.stopProject(id)
    showToast('Project stopped successfully', 'success')
    if (showDetailsModal.value && projectDetails.value?.id === id) {
      await loadProjectDetails(id)
    }
  } catch (err) {
    showToast('Failed to stop project', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleRestart(id) {
  actionLoading.value = id
  try {
    await composeStore.restartProject(id)
    showToast('Project restarted successfully', 'success')
    if (showDetailsModal.value && projectDetails.value?.id === id) {
      await loadProjectDetails(id)
    }
  } catch (err) {
    showToast('Failed to restart project', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleDetails(id) {
  showDetailsModal.value = true
  await loadProjectDetails(id)
}

async function loadProjectDetails(id) {
  loadingDetails.value = true
  projectDetails.value = null
  try {
    const details = await composeStore.getProject(id)
    projectDetails.value = details
  } catch (err) {
    showToast('Failed to load project details', 'error')
    showDetailsModal.value = false
  } finally {
    loadingDetails.value = false
  }
}

function closeDetailsModal() {
  showDetailsModal.value = false
  projectDetails.value = null
}

function handleDelete(id) {
  projectToDelete.value = id
  showDeleteModal.value = true
}

async function confirmDelete() {
  actionLoading.value = projectToDelete.value
  try {
    await composeStore.deleteProject(projectToDelete.value)
    showToast('Project deleted successfully', 'success')
    showDeleteModal.value = false
    if (showDetailsModal.value && projectDetails.value?.id === projectToDelete.value) {
      closeDetailsModal()
    }
  } catch (err) {
    showToast('Failed to delete project', 'error')
  } finally {
    actionLoading.value = null
    projectToDelete.value = null
  }
}

function getStatusVariant(status) {
  const s = status?.toLowerCase()
  if (s === 'running') return 'online'
  if (s === 'stopped') return 'warning'
  if (s === 'error') return 'offline'
  return 'warning'
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  fetchProjects()
})
</script>

<style scoped>
.compose-page {
  min-height: 100vh;
}

.compose-main {
  padding-top: var(--space-2xl);
  padding-bottom: var(--space-3xl);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-2xl);
}

.page-header h1 {
  margin-bottom: var(--space-sm);
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.header-actions {
  display: flex;
  gap: var(--space-md);
}

.tabs-section {
  display: flex;
  gap: var(--space-md);
  margin-bottom: var(--space-xl);
  border-bottom: 2px solid rgba(74, 85, 104, 0.3);
}

.tab-button {
  padding: var(--space-md) var(--space-lg);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-base);
  margin-bottom: -2px;
}

.tab-button:hover {
  color: var(--reach-amber);
}

.tab-button.active {
  color: var(--reach-amber);
  border-bottom-color: var(--reach-amber);
}

.tab-content {
  min-height: 400px;
}

.projects-grid,
.templates-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: var(--space-lg);
}

.loading-state,
.error-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-3xl);
  text-align: center;
  gap: var(--space-lg);
}

.error-message {
  color: var(--reach-orange);
  font-family: var(--font-mono);
}

.template-card {
  padding: var(--space-lg);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  transition: all var(--transition-base);
}

.template-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.template-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
}

.template-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
  flex: 1;
}

.deploy-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.details-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
  max-height: 70vh;
  overflow-y: auto;
}

.details-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.project-title {
  font-size: 1.25rem;
  color: var(--reach-amber);
  margin: 0 0 var(--space-xs) 0;
}

.project-desc {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin: 0;
}

.details-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.section-title {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--reach-amber);
  margin: 0;
}

.container-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.container-item {
  padding: var(--space-sm);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  color: var(--reach-cyan);
}

.details-actions {
  display: flex;
  gap: var(--space-sm);
}

.yaml-display {
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  padding: var(--space-md);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-primary);
  overflow-x: auto;
  max-height: 400px;
  white-space: pre-wrap;
}

.text-sm {
  font-size: 0.875rem;
}

.toast {
  position: fixed;
  bottom: var(--space-xl);
  right: var(--space-xl);
  padding: var(--space-md) var(--space-lg);
  background-color: var(--reach-steel);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  z-index: var(--z-toast);
  animation: slideIn 0.3s ease-out;
  font-family: var(--font-mono);
  font-size: 0.875rem;
}

.toast-success {
  border: 1px solid var(--reach-cyan);
  color: var(--reach-cyan);
}

.toast-error {
  border: 1px solid var(--reach-orange);
  color: var(--reach-orange);
}

.toast-warning {
  border: 1px solid var(--reach-amber);
  color: var(--reach-amber);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--space-md);
  }

  .projects-grid,
  .templates-grid {
    grid-template-columns: 1fr;
  }
}
</style>
