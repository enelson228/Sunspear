<template>
  <div class="images-page">
    <main class="images-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>IMAGE MANAGEMENT</h1>
            <p class="subtitle">Manage Docker images and registries</p>
          </div>
          <div class="header-actions">
            <Button variant="secondary" @click="showPruneModal = true">
              PRUNE
            </Button>
            <Button variant="secondary" @click="showBuildModal = true">
              BUILD
            </Button>
            <Button variant="primary" @click="fetchImages" :loading="loading">
              {{ loading ? 'REFRESHING...' : 'REFRESH' }}
            </Button>
            <Button variant="primary" @click="showPullModal = true">
              + PULL IMAGE
            </Button>
          </div>
        </div>

        <div class="search-section">
          <Input
            v-model="searchQuery"
            placeholder="Search images..."
            type="text"
          />
        </div>

        <div v-if="loading && images.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>Loading images...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="fetchImages">RETRY</Button>
        </div>

        <div v-else-if="filteredImages.length === 0" class="empty-state">
          <h3>NO IMAGES FOUND</h3>
          <p class="text-secondary">
            {{ searchQuery ? 'No images match your search.' : 'No images available. Pull an image to get started.' }}
          </p>
          <Button variant="primary" @click="showPullModal = true">
            PULL YOUR FIRST IMAGE
          </Button>
        </div>

        <div v-else class="images-grid">
          <ImageCard
            v-for="image in filteredImages"
            :key="image.Id"
            :image="image"
            :loading="actionLoading === image.Id"
            @create="handleCreateContainer"
            @remove="handleRemove"
            @tag="handleTagClick"
            @inspect="handleInspectClick"
          />
        </div>
      </div>
    </main>

    <!-- Pull Image Modal -->
    <Modal v-model="showPullModal" title="PULL IMAGE">
      <div class="pull-form">
        <Input
          v-model="pullImageName"
          label="IMAGE NAME"
          placeholder="nginx:latest"
          type="text"
        />
        <p class="text-secondary text-sm mt-sm">
          Examples: nginx:latest, redis:alpine, postgres:14
        </p>
      </div>

      <template #footer>
        <Button variant="secondary" @click="showPullModal = false">
          CANCEL
        </Button>
        <Button variant="primary" @click="handlePull" :loading="pulling">
          {{ pulling ? 'PULLING...' : 'PULL IMAGE' }}
        </Button>
      </template>
    </Modal>

    <!-- Remove Confirmation Modal -->
    <Modal v-model="showRemoveModal" title="CONFIRM REMOVAL">
      <p>Are you sure you want to remove this image?</p>
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

    <!-- Tag Image Modal -->
    <Modal v-model="showTagModal" title="TAG IMAGE">
      <div class="tag-form">
        <Input
          v-model="tagForm.repo"
          label="REPOSITORY"
          placeholder="myregistry/myimage"
          type="text"
        />
        <Input
          v-model="tagForm.tag"
          label="TAG"
          placeholder="v1.0.0"
          type="text"
        />
        <p class="text-secondary text-sm mt-sm">
          Example: myregistry/myimage:v1.0.0
        </p>
      </div>

      <template #footer>
        <Button variant="secondary" @click="showTagModal = false">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleTag" :loading="tagging">
          {{ tagging ? 'TAGGING...' : 'TAG IMAGE' }}
        </Button>
      </template>
    </Modal>

    <!-- Inspect Image Modal -->
    <Modal v-model="showInspectModal" title="IMAGE DETAILS">
      <div v-if="inspecting" class="loading-state">
        <div class="spinner"></div>
        <p>Loading image details...</p>
      </div>
      <div v-else-if="imageDetails" class="inspect-content">
        <div class="inspect-section">
          <h4 class="inspect-title">CONFIGURATION</h4>
          <pre class="inspect-code">{{ JSON.stringify(imageDetails.Config, null, 2) }}</pre>
        </div>
        <div class="inspect-section" v-if="imageHistory && imageHistory.length">
          <h4 class="inspect-title">LAYERS ({{ imageHistory.length }})</h4>
          <div class="layers-list">
            <div v-for="(layer, index) in imageHistory" :key="index" class="layer-item">
              <span class="layer-id font-mono">{{ layer.Id?.substring(0, 12) || 'none' }}</span>
              <span class="layer-size">{{ formatBytes(layer.Size) }}</span>
              <span class="layer-created">{{ formatDate(layer.Created) }}</span>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeInspectModal">
          CLOSE
        </Button>
      </template>
    </Modal>

    <!-- Prune Images Modal -->
    <Modal v-model="showPruneModal" title="CONFIRM PRUNE">
      <p>Are you sure you want to prune unused images?</p>
      <p class="text-secondary mt-sm">This will remove all dangling images not used by any container. This action cannot be undone.</p>

      <template #footer>
        <Button variant="secondary" @click="showPruneModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handlePrune" :loading="pruning">
          {{ pruning ? 'PRUNING...' : 'PRUNE IMAGES' }}
        </Button>
      </template>
    </Modal>

    <!-- Build Image Modal -->
    <Modal v-model="showBuildModal" title="BUILD IMAGE">
      <div class="build-form">
        <Input
          v-model="buildForm.tags"
          label="IMAGE TAG *"
          placeholder="myimage:latest"
          type="text"
        />
        <div class="form-section">
          <label class="label">DOCKERFILE CONTENT *</label>
          <textarea
            v-model="buildForm.dockerfile"
            class="dockerfile-input"
            placeholder="FROM nginx:alpine&#10;COPY . /usr/share/nginx/html&#10;..."
            rows="12"
          ></textarea>
        </div>
      </div>

      <template #footer>
        <Button variant="secondary" @click="closeBuildModal">
          CANCEL
        </Button>
        <Button variant="primary" @click="handleBuild" :loading="building">
          {{ building ? 'BUILDING...' : 'BUILD IMAGE' }}
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
import { useImagesStore } from '@/stores/images'
import ImageCard from '@/components/images/ImageCard.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const imagesStore = useImagesStore()

const searchQuery = ref('')
const actionLoading = ref(null)
const showPullModal = ref(false)
const showRemoveModal = ref(false)
const showTagModal = ref(false)
const showInspectModal = ref(false)
const showPruneModal = ref(false)
const showBuildModal = ref(false)
const pullImageName = ref('')
const pulling = ref(false)
const tagging = ref(false)
const inspecting = ref(false)
const pruning = ref(false)
const building = ref(false)
const imageToRemove = ref(null)
const imageToTag = ref(null)
const imageDetails = ref(null)
const imageHistory = ref(null)
const tagForm = ref({ repo: '', tag: '' })
const buildForm = ref({ tags: '', dockerfile: '' })
const toast = ref({ show: false, message: '', type: 'success' })

const images = computed(() => imagesStore.images)
const loading = computed(() => imagesStore.loading)
const error = computed(() => imagesStore.error)

const filteredImages = computed(() => {
  if (!searchQuery.value) return images.value

  const query = searchQuery.value.toLowerCase()
  return images.value.filter(image => {
    const tags = image.RepoTags?.join(' ').toLowerCase() || ''
    const id = image.Id.toLowerCase()
    return tags.includes(query) || id.includes(query)
  })
})

async function fetchImages() {
  try {
    await imagesStore.fetchImages()
  } catch (err) {
    showToast('Failed to fetch images', 'error')
  }
}

async function handlePull() {
  if (!pullImageName.value) {
    showToast('Please enter an image name', 'error')
    return
  }

  pulling.value = true
  try {
    await imagesStore.pullImage(pullImageName.value)
    showToast(`Successfully pulled ${pullImageName.value}`, 'success')
    showPullModal.value = false
    pullImageName.value = ''
  } catch (err) {
    showToast('Failed to pull image', 'error')
  } finally {
    pulling.value = false
  }
}

function handleRemove(id) {
  imageToRemove.value = id
  showRemoveModal.value = true
}

async function confirmRemove() {
  actionLoading.value = imageToRemove.value
  try {
    await imagesStore.removeImage(imageToRemove.value, false)
    showToast('Image removed successfully', 'success')
    showRemoveModal.value = false
  } catch (err) {
    showToast('Failed to remove image', 'error')
  } finally {
    actionLoading.value = null
    imageToRemove.value = null
  }
}

function handleCreateContainer(image) {
  router.push('/containers')
}

function handleTagClick(imageId) {
  imageToTag.value = imageId
  tagForm.value = { repo: '', tag: '' }
  showTagModal.value = true
}

async function handleTag() {
  if (!tagForm.value.repo) {
    showToast('Repository name is required', 'error')
    return
  }
  tagging.value = true
  try {
    await imagesStore.tagImage(imageToTag.value, tagForm.value.repo, tagForm.value.tag || 'latest')
    showToast('Image tagged successfully', 'success')
    showTagModal.value = false
  } catch (err) {
    showToast('Failed to tag image', 'error')
  } finally {
    tagging.value = false
  }
}

async function handleInspectClick(imageId) {
  showInspectModal.value = true
  inspecting.value = true
  imageDetails.value = null
  imageHistory.value = null
  try {
    const [details, history] = await Promise.all([
      imagesStore.inspectImage(imageId),
      imagesStore.getImageHistory(imageId)
    ])
    imageDetails.value = details
    imageHistory.value = history
  } catch (err) {
    showToast('Failed to load image details', 'error')
    showInspectModal.value = false
  } finally {
    inspecting.value = false
  }
}

function closeInspectModal() {
  showInspectModal.value = false
  imageDetails.value = null
  imageHistory.value = null
}

async function handlePrune() {
  pruning.value = true
  try {
    const result = await imagesStore.pruneImages()
    const space = formatBytes(result.spaceReclaimed || 0)
    const count = result.deleted?.length || 0
    showToast(`Pruned ${count} image(s). Space reclaimed: ${space}`, 'success')
    showPruneModal.value = false
  } catch (err) {
    showToast('Failed to prune images', 'error')
  } finally {
    pruning.value = false
  }
}

async function handleBuild() {
  if (!buildForm.value.tags || !buildForm.value.dockerfile) {
    showToast('Image tag and Dockerfile content are required', 'error')
    return
  }
  building.value = true
  try {
    await imagesStore.buildImage(buildForm.value.dockerfile, buildForm.value.tags)
    showToast('Image built successfully', 'success')
    closeBuildModal()
  } catch (err) {
    showToast('Failed to build image', 'error')
  } finally {
    building.value = false
  }
}

function closeBuildModal() {
  showBuildModal.value = false
  buildForm.value = { tags: '', dockerfile: '' }
}

function formatBytes(bytes) {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i]
}

function formatDate(timestamp) {
  if (!timestamp) return 'N/A'
  return new Date(timestamp * 1000).toISOString().replace('T', ' ').substring(0, 19)
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  fetchImages()
})
</script>

<style scoped>
.images-page {
  min-height: 100vh;
}

.images-main {
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

.images-grid {
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

.pull-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
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

.tag-form,
.build-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
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

.layers-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.layer-item {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: var(--space-md);
  padding: var(--space-sm);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.layer-id {
  color: var(--reach-cyan);
}

.layer-size {
  color: var(--text-secondary);
}

.layer-created {
  color: var(--text-muted);
  font-size: 0.7rem;
}

.dockerfile-input {
  width: 100%;
  min-height: 300px;
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  resize: vertical;
}

.dockerfile-input:focus {
  outline: none;
  border-color: var(--reach-amber);
  box-shadow: 0 0 0 2px rgba(246, 166, 35, 0.2);
}

.dockerfile-input::placeholder {
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--space-md);
  }

  .images-grid {
    grid-template-columns: 1fr;
  }
}
</style>
