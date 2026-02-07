<template>
  <div class="appstore-page">
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

    <main class="appstore-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>APP STORE</h1>
            <p class="subtitle">Search Docker Hub and pull images</p>
          </div>
        </div>

        <div class="search-section">
          <Input
            v-model="searchTerm"
            placeholder="Search Docker images (e.g., nginx, postgres)..."
            type="text"
            @keyup.enter="handleSearch"
          />
          <Button variant="primary" @click="handleSearch" :loading="searching">
            {{ searching ? 'SEARCHING...' : 'SEARCH' }}
          </Button>
        </div>

        <div v-if="searching" class="loading-state">
          <div class="spinner"></div>
          <p>Searching registry...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
        </div>

        <div v-else-if="results.length === 0" class="empty-state">
          <h3>NO RESULTS</h3>
          <p class="text-secondary">Search for a Docker image to see results.</p>
        </div>

        <div v-else class="results-grid">
          <div v-for="result in results" :key="result.Name" class="result-card hud-corners">
            <div class="result-header">
              <div>
                <h3>{{ result.Name }}</h3>
                <p class="text-secondary">{{ result.Description || 'No description available.' }}</p>
              </div>
              <div class="badges">
                <span v-if="result.IsOfficial" class="badge badge-official">OFFICIAL</span>
                <span v-if="result.IsAutomated" class="badge badge-automated">AUTOMATED</span>
              </div>
            </div>

            <div class="result-meta">
              <span class="meta-item">⭐ {{ result.StarCount || 0 }}</span>
            </div>

            <div class="result-actions">
              <Button variant="primary" @click="openPull(result.Name)">
                PULL IMAGE
              </Button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <Modal v-model="showPullModal" title="PULL IMAGE">
      <div class="pull-form">
        <Input
          v-model="pullImageName"
          label="IMAGE NAME"
          placeholder="nginx:latest"
          type="text"
        />
        <p class="text-secondary text-sm mt-sm">
          Tip: add a tag like `:latest` or a version.
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

    <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useImagesStore } from '@/stores/images'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'

const router = useRouter()
const authStore = useAuthStore()
const imagesStore = useImagesStore()

const searchTerm = ref('')
const searching = ref(false)
const results = ref([])
const error = ref('')

const showPullModal = ref(false)
const pullImageName = ref('')
const pulling = ref(false)

const toast = ref({ show: false, message: '', type: 'success' })

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

async function handleSearch() {
  if (!searchTerm.value.trim()) {
    error.value = 'Enter a search term.'
    results.value = []
    return
  }

  error.value = ''
  searching.value = true
  try {
    results.value = await imagesStore.searchImages(searchTerm.value.trim())
  } catch (err) {
    error.value = 'Search failed. Please try again.'
  } finally {
    searching.value = false
  }
}

function openPull(name) {
  pullImageName.value = name
  showPullModal.value = true
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
  } catch (err) {
    showToast('Failed to pull image', 'error')
  } finally {
    pulling.value = false
  }
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}
</script>

<style scoped>
.appstore-page {
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

.appstore-main {
  padding-top: 100px;
  padding-bottom: var(--space-3xl);
}

.page-header {
  margin-bottom: var(--space-2xl);
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.search-section {
  display: flex;
  gap: var(--space-md);
  align-items: center;
  margin-bottom: var(--space-2xl);
}

.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-lg);
}

.result-card {
  padding: var(--space-lg);
  background-color: var(--reach-steel);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-lg);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.result-header h3 {
  margin-bottom: var(--space-xs);
}

.badges {
  display: flex;
  gap: var(--space-xs);
  margin-top: var(--space-sm);
}

.badge {
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.65rem;
  letter-spacing: 0.12em;
}

.badge-official {
  border: 1px solid rgba(34, 211, 238, 0.6);
  color: var(--reach-cyan);
}

.badge-automated {
  border: 1px solid rgba(246, 166, 35, 0.6);
  color: var(--reach-amber);
}

.result-meta {
  display: flex;
  gap: var(--space-md);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.result-actions {
  margin-top: auto;
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

@media (max-width: 768px) {
  .search-section {
    flex-direction: column;
    align-items: stretch;
  }

  .results-grid {
    grid-template-columns: 1fr;
  }
}
</style>
