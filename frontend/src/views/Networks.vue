<template>
  <div class="networks-page">
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
            <router-link to="/volumes" class="nav-link">Volumes</router-link>
            <router-link to="/networks" class="nav-link">Networks</router-link>
            <router-link to="/compose" class="nav-link">Compose</router-link>
            <router-link to="/apps" class="nav-link">App Store</router-link>
            <router-link to="/system" class="nav-link">System</router-link>
          </div>
          <Button variant="secondary" @click="handleLogout">LOGOUT</Button>
        </div>
      </div>
    </nav>

    <main class="networks-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>NETWORK MANAGEMENT</h1>
            <p class="subtitle">Manage Docker networks</p>
          </div>
          <div class="header-actions">
            <Button variant="secondary" @click="showPruneModal = true">
              PRUNE
            </Button>
            <Button variant="primary" @click="fetchNetworks" :loading="loading">
              {{ loading ? 'REFRESHING...' : 'REFRESH' }}
            </Button>
            <Button variant="primary" @click="showCreateModal = true">
              + CREATE NETWORK
            </Button>
          </div>
        </div>

        <div class="search-section">
          <Input
            v-model="searchQuery"
            placeholder="Search networks..."
            type="text"
          />
        </div>

        <div v-if="loading && networks.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>Loading networks...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="fetchNetworks">RETRY</Button>
        </div>

        <div v-else-if="filteredNetworks.length === 0" class="empty-state">
          <h3>NO NETWORKS FOUND</h3>
          <p class="text-secondary">
            {{ searchQuery ? 'No networks match your search.' : 'No networks available. Create a network to get started.' }}
          </p>
          <Button variant="primary" @click="showCreateModal = true">
            CREATE YOUR FIRST NETWORK
          </Button>
        </div>

        <div v-else class="networks-grid">
          <NetworkCard
            v-for="network in filteredNetworks"
            :key="network.Id"
            :network="network"
            :loading="actionLoading === network.Id"
            @inspect="handleInspectClick"
            @remove="handleRemove"
          />
        </div>
      </div>
    </main>

    <!-- Create Network Modal -->
    <Modal v-model="showCreateModal" title="CREATE NETWORK">
      <div class="create-form">
        <Input
          v-model="createForm.name"
          label="NETWORK NAME *"
          placeholder="my-network"
          type="text"
        />
        <div class="form-section">
          <label class="label">DRIVER</label>
          <select v-model="createForm.driver" class="driver-select">
            <option value="bridge">bridge</option>
            <option value="overlay">overlay</option>
            <option value="macvlan">macvlan</option>
            <option value="host">host</option>
          </select>
        </div>
        <div class="form-section">
          <label class="checkbox-label">
            <input type="checkbox" v-model="createForm.internal" />
            <span>INTERNAL NETWORK (restrict external access)</span>
          </label>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeCreateModal">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleCreate" :loading="creating">
          {{ creating ? 'CREATING...' : 'CREATE NETWORK' }}
        </Button>
      </template>
    </Modal>

    <!-- Inspect Network Modal -->
    <Modal v-model="showInspectModal" title="NETWORK DETAILS">
      <div v-if="inspecting" class="loading-state">
        <div class="spinner"></div>
        <p>Loading network details...</p>
      </div>
      <div v-else-if="networkDetails" class="inspect-content">
        <div class="inspect-section">
          <h4 class="inspect-title">IPAM CONFIGURATION</h4>
          <div v-if="networkDetails.IPAM?.Config?.length" class="ipam-list">
            <div v-for="(config, index) in networkDetails.IPAM.Config" :key="index" class="ipam-item">
              <div class="detail-row">
                <span class="label">SUBNET</span>
                <span class="value font-mono">{{ config.Subnet || 'N/A' }}</span>
              </div>
              <div class="detail-row">
                <span class="label">GATEWAY</span>
                <span class="value font-mono">{{ config.Gateway || 'N/A' }}</span>
              </div>
            </div>
          </div>
          <p v-else class="text-secondary">No IPAM configuration</p>
        </div>

        <div class="inspect-section" v-if="getConnectedContainers().length">
          <h4 class="inspect-title">CONNECTED CONTAINERS ({{ getConnectedContainers().length }})</h4>
          <div class="containers-table">
            <div v-for="container in getConnectedContainers()" :key="container.id" class="container-row">
              <div class="container-info">
                <span class="container-name">{{ container.name }}</span>
                <span class="container-id font-mono">{{ container.id }}</span>
              </div>
              <Button
                variant="danger"
                size="sm"
                @click="handleDisconnect(container.id)"
                :loading="disconnecting === container.id"
              >
                DISCONNECT
              </Button>
            </div>
          </div>
        </div>

        <div class="inspect-section">
          <h4 class="inspect-title">CONNECT CONTAINER</h4>
          <div class="connect-form">
            <Input
              v-model="connectContainerId"
              placeholder="Container ID or Name"
              type="text"
            />
            <Button
              variant="primary"
              @click="handleConnect"
              :loading="connecting"
            >
              {{ connecting ? 'CONNECTING...' : 'CONNECT' }}
            </Button>
          </div>
        </div>

        <div class="inspect-section">
          <h4 class="inspect-title">FULL DETAILS</h4>
          <pre class="inspect-code">{{ JSON.stringify(networkDetails, null, 2) }}</pre>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeInspectModal">
          CLOSE
        </Button>
      </template>
    </Modal>

    <!-- Remove Confirmation Modal -->
    <Modal v-model="showRemoveModal" title="CONFIRM REMOVAL">
      <p>Are you sure you want to remove this network?</p>
      <p class="text-secondary mt-sm" v-if="isBuiltInNetwork()">
        WARNING: This is a built-in Docker network. Removing it may cause issues.
      </p>
      <p class="text-secondary mt-sm" v-else>
        This action cannot be undone.
      </p>

      <template #footer>
        <Button variant="secondary" @click="showRemoveModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="confirmRemove" :loading="actionLoading">
          {{ actionLoading ? 'REMOVING...' : 'REMOVE' }}
        </Button>
      </template>
    </Modal>

    <!-- Prune Networks Modal -->
    <Modal v-model="showPruneModal" title="CONFIRM PRUNE">
      <p>Are you sure you want to prune unused networks?</p>
      <p class="text-secondary mt-sm">This will remove all networks not used by any container. This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showPruneModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handlePrune" :loading="pruning">
          {{ pruning ? 'PRUNING...' : 'PRUNE NETWORKS' }}
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
import { useNetworksStore } from '@/stores/networks'
import NetworkCard from '@/components/networks/NetworkCard.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const authStore = useAuthStore()
const networksStore = useNetworksStore()

const searchQuery = ref('')
const actionLoading = ref(null)
const showCreateModal = ref(false)
const showRemoveModal = ref(false)
const showInspectModal = ref(false)
const showPruneModal = ref(false)
const creating = ref(false)
const inspecting = ref(false)
const pruning = ref(false)
const connecting = ref(false)
const disconnecting = ref(null)
const networkToRemove = ref(null)
const networkDetails = ref(null)
const connectContainerId = ref('')
const createForm = ref({ name: '', driver: 'bridge', internal: false })
const toast = ref({ show: false, message: '', type: 'success' })

const networks = computed(() => networksStore.networks)
const loading = computed(() => networksStore.loading)
const error = computed(() => networksStore.error)

const filteredNetworks = computed(() => {
  if (!searchQuery.value) return networks.value

  const query = searchQuery.value.toLowerCase()
  return networks.value.filter(network => {
    return network.Name.toLowerCase().includes(query) || network.Id.toLowerCase().includes(query)
  })
})

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

async function fetchNetworks() {
  try {
    await networksStore.fetchNetworks()
  } catch (err) {
    showToast('Failed to fetch networks', 'error')
  }
}

async function handleCreate() {
  if (!createForm.value.name) {
    showToast('Network name is required', 'error')
    return
  }

  creating.value = true
  try {
    await networksStore.createNetwork({
      name: createForm.value.name,
      driver: createForm.value.driver,
      internal: createForm.value.internal
    })
    showToast(`Network ${createForm.value.name} created successfully`, 'success')
    closeCreateModal()
  } catch (err) {
    showToast('Failed to create network', 'error')
  } finally {
    creating.value = false
  }
}

function closeCreateModal() {
  showCreateModal.value = false
  createForm.value = { name: '', driver: 'bridge', internal: false }
}

function handleRemove(id) {
  networkToRemove.value = id
  showRemoveModal.value = true
}

function isBuiltInNetwork() {
  if (!networkDetails.value) return false
  const builtInNetworks = ['bridge', 'host', 'none']
  return builtInNetworks.includes(networkDetails.value.Name)
}

async function confirmRemove() {
  actionLoading.value = networkToRemove.value
  try {
    await networksStore.removeNetwork(networkToRemove.value)
    showToast('Network removed successfully', 'success')
    showRemoveModal.value = false
  } catch (err) {
    showToast('Failed to remove network', 'error')
  } finally {
    actionLoading.value = null
    networkToRemove.value = null
  }
}

async function handleInspectClick(id) {
  showInspectModal.value = true
  inspecting.value = true
  networkDetails.value = null
  connectContainerId.value = ''
  try {
    const details = await networksStore.inspectNetwork(id)
    networkDetails.value = details
  } catch (err) {
    showToast('Failed to load network details', 'error')
    showInspectModal.value = false
  } finally {
    inspecting.value = false
  }
}

function getConnectedContainers() {
  if (!networkDetails.value?.Containers) return []
  return Object.entries(networkDetails.value.Containers).map(([id, info]) => ({
    id,
    name: info.Name || 'Unknown'
  }))
}

async function handleConnect() {
  if (!connectContainerId.value) {
    showToast('Container ID is required', 'error')
    return
  }

  connecting.value = true
  try {
    await networksStore.connectContainer(networkDetails.value.Id, connectContainerId.value)
    showToast('Container connected successfully', 'success')
    connectContainerId.value = ''
    const details = await networksStore.inspectNetwork(networkDetails.value.Id)
    networkDetails.value = details
  } catch (err) {
    showToast('Failed to connect container', 'error')
  } finally {
    connecting.value = false
  }
}

async function handleDisconnect(containerId) {
  disconnecting.value = containerId
  try {
    await networksStore.disconnectContainer(networkDetails.value.Id, containerId)
    showToast('Container disconnected successfully', 'success')
    const details = await networksStore.inspectNetwork(networkDetails.value.Id)
    networkDetails.value = details
  } catch (err) {
    showToast('Failed to disconnect container', 'error')
  } finally {
    disconnecting.value = null
  }
}

function closeInspectModal() {
  showInspectModal.value = false
  networkDetails.value = null
  connectContainerId.value = ''
}

async function handlePrune() {
  pruning.value = true
  try {
    const result = await networksStore.pruneNetworks()
    const count = result.deleted?.length || 0
    showToast(`Pruned ${count} network(s)`, 'success')
    showPruneModal.value = false
  } catch (err) {
    showToast('Failed to prune networks', 'error')
  } finally {
    pruning.value = false
  }
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  fetchNetworks()
})
</script>

<style scoped>
.networks-page {
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

.networks-main {
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

.search-section {
  margin-bottom: var(--space-xl);
  max-width: 500px;
}

.networks-grid {
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

.create-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.label {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.driver-select {
  width: 100%;
  padding: var(--space-sm) var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
}

.driver-select:focus {
  outline: none;
  border-color: var(--reach-amber);
  box-shadow: 0 0 0 2px rgba(246, 166, 35, 0.2);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.inspect-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
  max-height: 60vh;
  overflow-y: auto;
}

.inspect-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.inspect-title {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--reach-amber);
  margin: 0;
}

.ipam-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.ipam-item {
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.detail-row + .detail-row {
  margin-top: var(--space-sm);
}

.detail-row .label {
  min-width: 100px;
}

.detail-row .value {
  color: var(--text-primary);
  text-align: right;
  flex: 1;
}

.containers-table {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.container-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
  gap: var(--space-md);
}

.container-info {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  flex: 1;
}

.container-name {
  font-size: 0.875rem;
  color: var(--text-primary);
}

.container-id {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.connect-form {
  display: flex;
  gap: var(--space-md);
  align-items: flex-end;
}

.inspect-code {
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

  .networks-grid {
    grid-template-columns: 1fr;
  }
}
</style>
