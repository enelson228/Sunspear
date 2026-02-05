<template>
  <div class="images-page">
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

    <main class="images-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>IMAGE MANAGEMENT</h1>
            <p class="subtitle">Manage Docker images and registries</p>
          </div>
          <div class="header-actions">
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
import { useImagesStore } from '@/stores/images'
import ImageCard from '@/components/images/ImageCard.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const authStore = useAuthStore()
const imagesStore = useImagesStore()

const searchQuery = ref('')
const actionLoading = ref(null)
const showPullModal = ref(false)
const showRemoveModal = ref(false)
const pullImageName = ref('')
const pulling = ref(false)
const imageToRemove = ref(null)
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

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

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
  showToast('Container creation from image coming soon', 'warning')
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

.images-main {
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
