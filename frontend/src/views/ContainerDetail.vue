<template>
  <div class="container-detail-page">
    <nav class="navbar glass">
      <div class="container">
        <div class="navbar-content">
          <div class="logo-section">
            <div class="logo-icon">◆</div>
            <span class="logo-text">SUNSPEAR</span>
          </div>
          <div class="nav-links">
            <router-link to="/" class="nav-link">Dashboard</router-link>
            <router-link to="/containers" class="nav-link">Containers</router-link>
            <router-link to="/images" class="nav-link">Images</router-link>
            <router-link to="/apps" class="nav-link">App Store</router-link>
            <router-link to="/system" class="nav-link">System</router-link>
          </div>
          <Button variant="secondary" @click="handleLogout">LOGOUT</Button>
        </div>
      </div>
    </nav>

    <main class="detail-main">
      <div class="container">
        <div class="page-header">
          <div>
            <Button variant="secondary" @click="$router.push('/containers')" class="mb-md">
              ← BACK TO CONTAINERS
            </Button>
            <h1>CONTAINER DETAILS</h1>
            <p class="subtitle font-mono">{{ containerName }}</p>
          </div>
          <div class="header-actions" v-if="containerData">
            <Badge :variant="statusVariant" :pulse="containerData.State.Running">
              {{ containerData.State.Status }}
            </Badge>
          </div>
        </div>

        <div v-if="loading && !containerData" class="loading-state">
          <div class="spinner"></div>
          <p>Loading container details...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="fetchContainer">RETRY</Button>
        </div>

        <div v-else-if="containerData" class="detail-content">
          <!-- Action Buttons -->
          <Card class="actions-section">
            <h3 class="section-title">ACTIONS</h3>
            <div class="action-buttons">
              <Button
                v-if="!containerData.State.Running"
                variant="primary"
                :loading="actionLoading === 'start'"
                @click="handleStart"
              >
                START CONTAINER
              </Button>
              <Button
                v-if="containerData.State.Running"
                variant="secondary"
                :loading="actionLoading === 'stop'"
                @click="handleStop"
              >
                STOP CONTAINER
              </Button>
              <Button
                v-if="containerData.State.Running"
                variant="secondary"
                :loading="actionLoading === 'restart'"
                @click="handleRestart"
              >
                RESTART CONTAINER
              </Button>
              <Button
                variant="danger"
                :loading="actionLoading === 'remove'"
                @click="showRemoveModal = true"
              >
                REMOVE CONTAINER
              </Button>
            </div>
          </Card>

          <!-- Container Information -->
          <div class="info-grid">
            <Card class="info-card">
              <h3 class="section-title">GENERAL</h3>
              <div class="info-rows">
                <div class="info-row">
                  <span class="info-label">ID</span>
                  <span class="info-value font-mono">{{ containerData.Id.substring(0, 12) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">NAME</span>
                  <span class="info-value">{{ containerName }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">IMAGE</span>
                  <span class="info-value font-mono">{{ containerData.Config.Image }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">CREATED</span>
                  <span class="info-value">{{ formatDate(containerData.Created) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">STATUS</span>
                  <span class="info-value">{{ containerData.State.Status }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">RESTART POLICY</span>
                  <span class="info-value">{{ containerData.HostConfig.RestartPolicy.Name || 'none' }}</span>
                </div>
              </div>
            </Card>

            <Card class="info-card">
              <h3 class="section-title">NETWORK</h3>
              <div class="info-rows">
                <div class="info-row">
                  <span class="info-label">IP ADDRESS</span>
                  <span class="info-value font-mono">{{ getIPAddress() }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">PORTS</span>
                  <span class="info-value font-mono">{{ formatPorts() }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">NETWORKS</span>
                  <span class="info-value">{{ getNetworks() }}</span>
                </div>
              </div>
            </Card>

            <Card class="info-card">
              <h3 class="section-title">STATE</h3>
              <div class="info-rows">
                <div class="info-row">
                  <span class="info-label">RUNNING</span>
                  <span class="info-value">{{ containerData.State.Running ? 'Yes' : 'No' }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">PAUSED</span>
                  <span class="info-value">{{ containerData.State.Paused ? 'Yes' : 'No' }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">RESTARTING</span>
                  <span class="info-value">{{ containerData.State.Restarting ? 'Yes' : 'No' }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">EXIT CODE</span>
                  <span class="info-value">{{ containerData.State.ExitCode }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">STARTED AT</span>
                  <span class="info-value">{{ formatDate(containerData.State.StartedAt) }}</span>
                </div>
              </div>
            </Card>
          </div>

          <!-- Environment Variables -->
          <Card class="env-card" v-if="containerData.Config.Env && containerData.Config.Env.length">
            <h3 class="section-title">ENVIRONMENT VARIABLES</h3>
            <div class="env-list">
              <div v-for="(env, index) in containerData.Config.Env" :key="index" class="env-item font-mono">
                {{ env }}
              </div>
            </div>
          </Card>

          <!-- Logs -->
          <LogViewer
            :logs="logs"
            :loading="logsLoading"
            @refresh="fetchLogs"
            @download="downloadLogs"
          />
        </div>
      </div>
    </main>

    <!-- Remove Confirmation Modal -->
    <Modal v-model="showRemoveModal" title="CONFIRM REMOVAL">
      <p>Are you sure you want to remove container <strong>{{ containerName }}</strong>?</p>
      <p class="text-secondary mt-sm">This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showRemoveModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handleRemove" :loading="actionLoading === 'remove'">
          {{ actionLoading === 'remove' ? 'REMOVING...' : 'REMOVE' }}
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
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useContainersStore } from '@/stores/containers'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import Modal from '@/components/ui/Modal.vue'
import LogViewer from '@/components/containers/LogViewer.vue'
import dayjs from 'dayjs'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const containersStore = useContainersStore()

const containerData = ref(null)
const logs = ref('')
const loading = ref(false)
const logsLoading = ref(false)
const error = ref(null)
const actionLoading = ref(null)
const showRemoveModal = ref(false)
const toast = ref({ show: false, message: '', type: 'success' })

const containerId = computed(() => route.params.id)
const containerName = computed(() => {
  if (!containerData.value) return ''
  return containerData.value.Name?.replace('/', '') || 'Unknown'
})

const statusVariant = computed(() => {
  if (!containerData.value) return 'offline'
  if (containerData.value.State.Running) return 'online'
  if (containerData.value.State.Paused) return 'warning'
  return 'offline'
})

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

async function fetchContainer() {
  loading.value = true
  error.value = null
  try {
    containerData.value = await containersStore.getContainer(containerId.value)
    await fetchLogs()
  } catch (err) {
    error.value = err.message || 'Failed to fetch container details'
  } finally {
    loading.value = false
  }
}

async function fetchLogs() {
  logsLoading.value = true
  try {
    logs.value = await containersStore.getLogs(containerId.value, 200)
  } catch (err) {
    console.error('Failed to fetch logs:', err)
  } finally {
    logsLoading.value = false
  }
}

async function handleStart() {
  actionLoading.value = 'start'
  try {
    await containersStore.startContainer(containerId.value)
    showToast('Container started successfully', 'success')
    await fetchContainer()
  } catch (err) {
    showToast('Failed to start container', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleStop() {
  actionLoading.value = 'stop'
  try {
    await containersStore.stopContainer(containerId.value)
    showToast('Container stopped successfully', 'success')
    await fetchContainer()
  } catch (err) {
    showToast('Failed to stop container', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleRestart() {
  actionLoading.value = 'restart'
  try {
    await containersStore.restartContainer(containerId.value)
    showToast('Container restarted successfully', 'success')
    await fetchContainer()
  } catch (err) {
    showToast('Failed to restart container', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleRemove() {
  actionLoading.value = 'remove'
  try {
    await containersStore.removeContainer(containerId.value, true)
    showToast('Container removed successfully', 'success')
    showRemoveModal.value = false
    setTimeout(() => router.push('/containers'), 1000)
  } catch (err) {
    showToast('Failed to remove container', 'error')
  } finally {
    actionLoading.value = null
  }
}

function downloadLogs() {
  const blob = new Blob([logs.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${containerName.value}-logs.txt`
  a.click()
  URL.revokeObjectURL(url)
}

function formatDate(dateString) {
  if (!dateString) return 'N/A'
  return dayjs(dateString).format('YYYY-MM-DD HH:mm:ss')
}

function getIPAddress() {
  if (!containerData.value?.NetworkSettings?.Networks) return 'N/A'
  const networks = Object.values(containerData.value.NetworkSettings.Networks)
  return networks[0]?.IPAddress || 'N/A'
}

function formatPorts() {
  if (!containerData.value?.NetworkSettings?.Ports) return 'None'
  const ports = Object.entries(containerData.value.NetworkSettings.Ports)
    .map(([internal, external]) => {
      if (external && external.length > 0) {
        return `${external[0].HostPort}:${internal}`
      }
      return internal
    })
    .join(', ')
  return ports || 'None'
}

function getNetworks() {
  if (!containerData.value?.NetworkSettings?.Networks) return 'None'
  return Object.keys(containerData.value.NetworkSettings.Networks).join(', ')
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  fetchContainer()
})
</script>

<style scoped>
.container-detail-page {
  min-height: 100vh;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  height: 64px;
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.navbar-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
}

.logo-section {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.logo-icon {
  color: var(--reach-amber);
  font-size: 1.5rem;
}

.logo-text {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: var(--reach-amber);
}

.nav-links {
  display: flex;
  gap: var(--space-xl);
}

.nav-link {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  transition: color var(--transition-base);
  position: relative;
}

.nav-link:hover {
  color: var(--reach-amber);
}

.nav-link.router-link-active {
  color: var(--reach-amber);
}

.nav-link.router-link-active::after {
  content: '';
  position: absolute;
  bottom: -20px;
  left: 0;
  right: 0;
  height: 2px;
  background-color: var(--reach-amber);
}

.detail-main {
  padding-top: 100px;
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
  color: var(--reach-cyan);
}

.header-actions {
  display: flex;
  gap: var(--space-md);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

.actions-section {
  padding: var(--space-lg);
}

.section-title {
  font-size: 1rem;
  margin-bottom: var(--space-md);
  color: var(--reach-amber);
}

.action-buttons {
  display: flex;
  gap: var(--space-md);
  flex-wrap: wrap;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: var(--space-lg);
}

.info-card {
  padding: var(--space-lg);
}

.info-rows {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.info-row {
  display: flex;
  justify-content: space-between;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid rgba(74, 85, 104, 0.2);
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.info-value {
  color: var(--text-primary);
  text-align: right;
}

.env-card {
  padding: var(--space-lg);
}

.env-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  max-height: 300px;
  overflow-y: auto;
}

.env-item {
  padding: var(--space-sm);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  color: var(--text-primary);
}

.loading-state,
.error-state {
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
</style>
