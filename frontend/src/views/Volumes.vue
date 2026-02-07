<template>
  <div class="volumes-page">
    <main class="volumes-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>VOLUME MANAGEMENT</h1>
            <p class="subtitle">Manage Docker volumes and storage</p>
          </div>
          <div class="header-actions">
            <Button variant="secondary" @click="showPruneModal = true">
              PRUNE
            </Button>
            <Button variant="primary" @click="fetchVolumes" :loading="loading">
              {{ loading ? 'REFRESHING...' : 'REFRESH' }}
            </Button>
            <Button variant="primary" @click="showCreateModal = true">
              + CREATE VOLUME
            </Button>
          </div>
        </div>

        <div class="search-section">
          <Input
            v-model="searchQuery"
            placeholder="Search volumes..."
            type="text"
          />
        </div>

        <div v-if="loading && volumes.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>Loading volumes...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="fetchVolumes">RETRY</Button>
        </div>

        <div v-else-if="filteredVolumes.length === 0" class="empty-state">
          <h3>NO VOLUMES FOUND</h3>
          <p class="text-secondary">
            {{ searchQuery ? 'No volumes match your search.' : 'No volumes available. Create a volume to get started.' }}
          </p>
          <Button variant="primary" @click="showCreateModal = true">
            CREATE YOUR FIRST VOLUME
          </Button>
        </div>

        <div v-else class="volumes-grid">
          <VolumeCard
            v-for="volume in filteredVolumes"
            :key="volume.Name"
            :volume="volume"
            :loading="actionLoading === volume.Name"
            @inspect="handleInspectClick"
            @remove="handleRemove"
          />
        </div>
      </div>
    </main>

    <!-- Create Volume Modal -->
    <Modal v-model="showCreateModal" title="CREATE VOLUME">
      <div class="create-form">
        <Input
          v-model="createForm.name"
          label="VOLUME NAME *"
          placeholder="my-volume"
          type="text"
        />
        <Input
          v-model="createForm.driver"
          label="DRIVER"
          placeholder="local"
          type="text"
        />
        <div class="form-section">
          <label class="label">LABELS (KEY=VALUE PER LINE)</label>
          <textarea
            v-model="createForm.labels"
            class="labels-input"
            placeholder="environment=production&#10;team=backend"
            rows="4"
          ></textarea>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeCreateModal">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleCreate" :loading="creating">
          {{ creating ? 'CREATING...' : 'CREATE VOLUME' }}
        </Button>
      </template>
    </Modal>

    <!-- Inspect Volume Modal -->
    <Modal v-model="showInspectModal" title="VOLUME DETAILS">
      <div v-if="inspecting" class="loading-state">
        <div class="spinner"></div>
        <p>Loading volume details...</p>
      </div>
      <div v-else-if="volumeDetails" class="inspect-content">
        <div class="inspect-section">
          <h4 class="inspect-title">CONFIGURATION</h4>
          <pre class="inspect-code">{{ JSON.stringify(volumeDetails, null, 2) }}</pre>
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
      <p>Are you sure you want to remove this volume?</p>
      <p class="text-secondary mt-sm">This action cannot be undone. All data in this volume will be lost.</p>

      <template #footer>
        <Button variant="secondary" @click="showRemoveModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="confirmRemove" :loading="actionLoading">
          {{ actionLoading ? 'REMOVING...' : 'REMOVE' }}
        </Button>
      </template>
    </Modal>

    <!-- Prune Volumes Modal -->
    <Modal v-model="showPruneModal" title="CONFIRM PRUNE">
      <p>Are you sure you want to prune unused volumes?</p>
      <p class="text-secondary mt-sm">This will remove all volumes not used by any container. This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showPruneModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handlePrune" :loading="pruning">
          {{ pruning ? 'PRUNING...' : 'PRUNE VOLUMES' }}
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
import { useVolumesStore } from '@/stores/volumes'
import VolumeCard from '@/components/volumes/VolumeCard.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const volumesStore = useVolumesStore()

const searchQuery = ref('')
const actionLoading = ref(null)
const showCreateModal = ref(false)
const showRemoveModal = ref(false)
const showInspectModal = ref(false)
const showPruneModal = ref(false)
const creating = ref(false)
const inspecting = ref(false)
const pruning = ref(false)
const volumeToRemove = ref(null)
const volumeDetails = ref(null)
const createForm = ref({ name: '', driver: 'local', labels: '' })
const toast = ref({ show: false, message: '', type: 'success' })

const volumes = computed(() => volumesStore.volumes)
const loading = computed(() => volumesStore.loading)
const error = computed(() => volumesStore.error)

const filteredVolumes = computed(() => {
  if (!searchQuery.value) return volumes.value

  const query = searchQuery.value.toLowerCase()
  return volumes.value.filter(volume => {
    return volume.Name.toLowerCase().includes(query)
  })
})

async function fetchVolumes() {
  try {
    await volumesStore.fetchVolumes()
  } catch (err) {
    showToast('Failed to fetch volumes', 'error')
  }
}

async function handleCreate() {
  if (!createForm.value.name) {
    showToast('Volume name is required', 'error')
    return
  }

  creating.value = true
  try {
    const labels = {}
    if (createForm.value.labels) {
      createForm.value.labels.split('\n').forEach(line => {
        const [key, ...valueParts] = line.split('=')
        if (key && valueParts.length) {
          labels[key.trim()] = valueParts.join('=').trim()
        }
      })
    }

    await volumesStore.createVolume({
      name: createForm.value.name,
      driver: createForm.value.driver || 'local',
      labels
    })
    showToast(`Volume ${createForm.value.name} created successfully`, 'success')
    closeCreateModal()
  } catch (err) {
    showToast('Failed to create volume', 'error')
  } finally {
    creating.value = false
  }
}

function closeCreateModal() {
  showCreateModal.value = false
  createForm.value = { name: '', driver: 'local', labels: '' }
}

function handleRemove(name) {
  volumeToRemove.value = name
  showRemoveModal.value = true
}

async function confirmRemove() {
  actionLoading.value = volumeToRemove.value
  try {
    await volumesStore.removeVolume(volumeToRemove.value, false)
    showToast('Volume removed successfully', 'success')
    showRemoveModal.value = false
  } catch (err) {
    showToast('Failed to remove volume', 'error')
  } finally {
    actionLoading.value = null
    volumeToRemove.value = null
  }
}

async function handleInspectClick(name) {
  showInspectModal.value = true
  inspecting.value = true
  volumeDetails.value = null
  try {
    const details = await volumesStore.inspectVolume(name)
    volumeDetails.value = details
  } catch (err) {
    showToast('Failed to load volume details', 'error')
    showInspectModal.value = false
  } finally {
    inspecting.value = false
  }
}

function closeInspectModal() {
  showInspectModal.value = false
  volumeDetails.value = null
}

async function handlePrune() {
  pruning.value = true
  try {
    const result = await volumesStore.pruneVolumes()
    const count = result.deleted?.length || 0
    showToast(`Pruned ${count} volume(s)`, 'success')
    showPruneModal.value = false
  } catch (err) {
    showToast('Failed to prune volumes', 'error')
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
  fetchVolumes()
})
</script>

<style scoped>
.volumes-page {
  min-height: 100vh;
}

.volumes-main {
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

.search-section {
  margin-bottom: var(--space-xl);
  max-width: 500px;
}

.volumes-grid {
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

.labels-input {
  width: 100%;
  min-height: 100px;
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  resize: vertical;
}

.labels-input:focus {
  outline: none;
  border-color: var(--reach-amber);
  box-shadow: 0 0 0 2px rgba(246, 166, 35, 0.2);
}

.labels-input::placeholder {
  color: var(--text-muted);
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

  .volumes-grid {
    grid-template-columns: 1fr;
  }
}
</style>
