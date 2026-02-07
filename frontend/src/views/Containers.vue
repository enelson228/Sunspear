<template>
  <div class="containers-page">
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

    <main class="containers-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>CONTAINER MANAGEMENT</h1>
            <p class="subtitle">Manage and monitor Docker containers</p>
          </div>
          <div class="header-actions">
            <Button variant="secondary" @click="showBulkStopModal = true">
              STOP ALL
            </Button>
            <Button variant="secondary" @click="showBulkRestartModal = true">
              RESTART ALL
            </Button>
            <Button variant="primary" @click="fetchContainers" :loading="loading">
              {{ loading ? 'REFRESHING...' : 'REFRESH' }}
            </Button>
            <Button variant="primary" @click="showCreateModal = true">
              + CREATE CONTAINER
            </Button>
          </div>
        </div>

        <div class="filters-section">
          <div class="filter-group">
            <label class="label">FILTER</label>
            <div class="filter-buttons">
              <Button
                :variant="showAll ? 'primary' : 'secondary'"
                @click="showAll = true; fetchContainers(true)"
              >
                ALL ({{ containers.length }})
              </Button>
              <Button
                :variant="!showAll ? 'primary' : 'secondary'"
                @click="showAll = false; fetchContainers(false)"
              >
                RUNNING ({{ runningCount }})
              </Button>
            </div>
          </div>

          <div class="search-group">
            <Input
              v-model="searchQuery"
              placeholder="Search containers..."
              type="text"
            />
          </div>
        </div>

        <div v-if="loading && containers.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>Loading containers...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="fetchContainers">RETRY</Button>
        </div>

        <div v-else-if="filteredContainers.length === 0" class="empty-state">
          <h3>NO CONTAINERS FOUND</h3>
          <p class="text-secondary">
            {{ searchQuery ? 'No containers match your search.' : 'No containers available.' }}
          </p>
          <Button variant="primary" @click="showCreateModal = true">
            CREATE YOUR FIRST CONTAINER
          </Button>
        </div>

        <div v-else class="containers-grid">
          <ContainerCard
            v-for="container in filteredContainers"
            :key="container.Id"
            :container="container"
            :loading="actionLoading === container.Id"
            @start="handleStart"
            @stop="handleStop"
            @restart="handleRestart"
            @view="handleView"
            @remove="handleRemove"
          />
        </div>
      </div>
    </main>

    <!-- Remove Confirmation Modal -->
    <Modal v-model="showRemoveModal" title="CONFIRM REMOVAL">
      <p>Are you sure you want to remove this container?</p>
      <p class="text-secondary mt-sm">This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showRemoveModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="confirmRemove" :loading="actionLoading">
          {{ actionLoading ? 'REMOVING...' : 'REMOVE' }}
        </Button>
      </template>
    </Modal>

    <!-- Create Container Modal -->
    <Modal v-model="showCreateModal" title="CREATE CONTAINER">
      <div class="create-form">
        <Input
          v-model="createForm.image"
          label="IMAGE *"
          placeholder="nginx:latest"
          type="text"
        />
        <Input
          v-model="createForm.name"
          label="CONTAINER NAME"
          placeholder="my-container"
          type="text"
        />

        <div class="form-section">
          <div class="form-section-header">
            <label class="label">PORT MAPPINGS</label>
            <Button variant="secondary" size="sm" @click="addPort">+ ADD PORT</Button>
          </div>
          <div v-for="(port, index) in createForm.ports" :key="'port-'+index" class="dynamic-row">
            <Input v-model="port.host" placeholder="Host port" type="text" />
            <span class="row-separator">:</span>
            <Input v-model="port.container" placeholder="Container port" type="text" />
            <Button variant="danger" size="sm" @click="createForm.ports.splice(index, 1)">X</Button>
          </div>
        </div>

        <div class="form-section">
          <div class="form-section-header">
            <label class="label">VOLUME MOUNTS</label>
            <Button variant="secondary" size="sm" @click="addVolume">+ ADD VOLUME</Button>
          </div>
          <div v-for="(vol, index) in createForm.volumes" :key="'vol-'+index" class="dynamic-row">
            <Input v-model="vol.host" placeholder="Host path" type="text" />
            <span class="row-separator">:</span>
            <Input v-model="vol.container" placeholder="Container path" type="text" />
            <Button variant="danger" size="sm" @click="createForm.volumes.splice(index, 1)">X</Button>
          </div>
        </div>

        <div class="form-section">
          <div class="form-section-header">
            <label class="label">ENVIRONMENT VARIABLES</label>
            <Button variant="secondary" size="sm" @click="addEnv">+ ADD ENV</Button>
          </div>
          <div v-for="(env, index) in createForm.envVars" :key="'env-'+index" class="dynamic-row">
            <Input v-model="env.key" placeholder="KEY" type="text" />
            <span class="row-separator">=</span>
            <Input v-model="env.value" placeholder="VALUE" type="text" />
            <Button variant="danger" size="sm" @click="createForm.envVars.splice(index, 1)">X</Button>
          </div>
        </div>

        <div class="form-section">
          <label class="label">RESTART POLICY</label>
          <select v-model="createForm.restartPolicy" class="select-input">
            <option value="">No restart</option>
            <option value="always">Always</option>
            <option value="unless-stopped">Unless stopped</option>
            <option value="on-failure">On failure</option>
          </select>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeCreateModal">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleCreate" :loading="creating">
          {{ creating ? 'CREATING...' : 'CREATE CONTAINER' }}
        </Button>
      </template>
    </Modal>

    <!-- Bulk Stop Modal -->
    <Modal v-model="showBulkStopModal" title="CONFIRM STOP ALL">
      <p>Are you sure you want to stop all running containers?</p>
      <p class="text-secondary mt-sm">This will stop {{ runningCount }} running container(s).</p>

      <template #footer>
        <Button variant="secondary" @click="showBulkStopModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handleBulkStop" :loading="bulkLoading">
          {{ bulkLoading ? 'STOPPING...' : 'STOP ALL' }}
        </Button>
      </template>
    </Modal>

    <!-- Bulk Restart Modal -->
    <Modal v-model="showBulkRestartModal" title="CONFIRM RESTART ALL">
      <p>Are you sure you want to restart all running containers?</p>
      <p class="text-secondary mt-sm">This will restart {{ runningCount }} running container(s).</p>

      <template #footer>
        <Button variant="secondary" @click="showBulkRestartModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handleBulkRestart" :loading="bulkLoading">
          {{ bulkLoading ? 'RESTARTING...' : 'RESTART ALL' }}
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
import { useAuthStore } from '@/stores/auth'
import { useContainersStore } from '@/stores/containers'
import ContainerCard from '@/components/containers/ContainerCard.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const authStore = useAuthStore()
const containersStore = useContainersStore()

const showAll = ref(true)
const searchQuery = ref('')
const actionLoading = ref(null)
const showRemoveModal = ref(false)
const showCreateModal = ref(false)
const showBulkStopModal = ref(false)
const showBulkRestartModal = ref(false)
const containerToRemove = ref(null)
const creating = ref(false)
const bulkLoading = ref(false)
const toast = ref({ show: false, message: '', type: 'success' })
const createForm = ref({
  image: '',
  name: '',
  ports: [],
  volumes: [],
  envVars: [],
  restartPolicy: ''
})

const containers = computed(() => containersStore.containers)
const loading = computed(() => containersStore.loading)
const error = computed(() => containersStore.error)

const runningCount = computed(() => {
  return containers.value.filter(c => c.State === 'running').length
})

const filteredContainers = computed(() => {
  let filtered = containers.value

  if (!showAll.value) {
    filtered = filtered.filter(c => c.State === 'running')
  }

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(c => {
      const name = c.Names[0]?.toLowerCase() || ''
      const id = c.Id.toLowerCase()
      const image = c.Image.toLowerCase()
      return name.includes(query) || id.includes(query) || image.includes(query)
    })
  }

  return filtered
})

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

async function fetchContainers(all = true) {
  try {
    await containersStore.fetchContainers(all)
  } catch (err) {
    showToast('Failed to fetch containers', 'error')
  }
}

async function handleStart(id) {
  actionLoading.value = id
  try {
    await containersStore.startContainer(id)
    showToast('Container started successfully', 'success')
  } catch (err) {
    showToast('Failed to start container', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleStop(id) {
  actionLoading.value = id
  try {
    await containersStore.stopContainer(id)
    showToast('Container stopped successfully', 'success')
  } catch (err) {
    showToast('Failed to stop container', 'error')
  } finally {
    actionLoading.value = null
  }
}

async function handleRestart(id) {
  actionLoading.value = id
  try {
    await containersStore.restartContainer(id)
    showToast('Container restarted successfully', 'success')
  } catch (err) {
    showToast('Failed to restart container', 'error')
  } finally {
    actionLoading.value = null
  }
}

function handleView(id) {
  router.push(`/containers/${id}`)
}

function handleRemove(id) {
  containerToRemove.value = id
  showRemoveModal.value = true
}

async function confirmRemove() {
  actionLoading.value = containerToRemove.value
  try {
    await containersStore.removeContainer(containerToRemove.value, false)
    showToast('Container removed successfully', 'success')
    showRemoveModal.value = false
  } catch (err) {
    showToast('Failed to remove container. Try force removal.', 'error')
  } finally {
    actionLoading.value = null
    containerToRemove.value = null
  }
}

function addPort() {
  createForm.value.ports.push({ host: '', container: '' })
}

function addVolume() {
  createForm.value.volumes.push({ host: '', container: '' })
}

function addEnv() {
  createForm.value.envVars.push({ key: '', value: '' })
}

function closeCreateModal() {
  showCreateModal.value = false
  createForm.value = { image: '', name: '', ports: [], volumes: [], envVars: [], restartPolicy: '' }
}

async function handleCreate() {
  if (!createForm.value.image) {
    showToast('Image name is required', 'error')
    return
  }

  creating.value = true
  try {
    const ports = {}
    createForm.value.ports.forEach(p => {
      if (p.host && p.container) {
        ports[p.container + '/tcp'] = p.host
      }
    })

    const volumes = {}
    createForm.value.volumes.forEach(v => {
      if (v.host && v.container) {
        volumes[v.host] = v.container
      }
    })

    const env = createForm.value.envVars
      .filter(e => e.key)
      .map(e => `${e.key}=${e.value}`)

    await containersStore.createContainer({
      image: createForm.value.image,
      name: createForm.value.name,
      ports,
      volumes,
      env,
      restartPolicy: createForm.value.restartPolicy
    })
    showToast('Container created successfully', 'success')
    closeCreateModal()
  } catch (err) {
    showToast('Failed to create container', 'error')
  } finally {
    creating.value = false
  }
}

async function handleBulkStop() {
  bulkLoading.value = true
  try {
    const result = await containersStore.bulkStopContainers()
    showToast(`Stopped ${result.stopped} container(s)`, 'success')
    showBulkStopModal.value = false
  } catch (err) {
    showToast('Failed to stop containers', 'error')
  } finally {
    bulkLoading.value = false
  }
}

async function handleBulkRestart() {
  bulkLoading.value = true
  try {
    const result = await containersStore.bulkRestartContainers()
    showToast(`Restarted ${result.restarted} container(s)`, 'success')
    showBulkRestartModal.value = false
  } catch (err) {
    showToast('Failed to restart containers', 'error')
  } finally {
    bulkLoading.value = false
  }
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  fetchContainers(true)
})
</script>

<style scoped>
.containers-page {
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

.containers-main {
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

.filters-section {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: var(--space-lg);
  margin-bottom: var(--space-xl);
  padding: var(--space-lg);
  background-color: var(--reach-steel);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-md);
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.filter-buttons {
  display: flex;
  gap: var(--space-sm);
}

.search-group {
  flex: 1;
  max-width: 400px;
}

.containers-grid {
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

.create-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.form-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dynamic-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.dynamic-row .input-wrapper {
  flex: 1;
}

.row-separator {
  font-family: var(--font-mono);
  font-size: 1.25rem;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.select-input {
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
}

.select-input:focus {
  outline: none;
  border-color: var(--reach-amber);
  box-shadow: 0 0 0 2px rgba(246, 166, 35, 0.2);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--space-md);
  }

  .filters-section {
    flex-direction: column;
    align-items: stretch;
  }

  .search-group {
    max-width: none;
  }

  .containers-grid {
    grid-template-columns: 1fr;
  }
}
</style>
