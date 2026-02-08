<template>
  <div class="appstore-page">
    <main class="appstore-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>APP STORE</h1>
            <p class="subtitle">Deploy Applications with One Click</p>
          </div>
          <div class="installed-summary" v-if="installedApps.length > 0">
            <span class="summary-text">{{ installedApps.length }} Installed Apps</span>
            <Button variant="secondary" size="sm" @click="showInstalledModal = true">
              VIEW ALL
            </Button>
          </div>
        </div>

        <!-- Tabs -->
        <div class="tabs">
          <button
            :class="['tab', { active: activeTab === 'catalog' }]"
            @click="activeTab = 'catalog'"
          >
            CATALOG
          </button>
          <button
            :class="['tab', { active: activeTab === 'dockerhub' }]"
            @click="activeTab = 'dockerhub'"
          >
            DOCKER HUB
          </button>
        </div>

        <!-- CATALOG TAB -->
        <div v-show="activeTab === 'catalog'" class="tab-content">
          <!-- Category Filters -->
          <div class="filters-section">
            <button
              v-for="cat in categories"
              :key="cat.value"
              :class="['filter-btn', { active: selectedCategory === cat.value }]"
              @click="selectedCategory = cat.value"
            >
              {{ cat.label }}
            </button>
          </div>

          <!-- Search -->
          <div class="search-section">
            <Input
              v-model="catalogSearch"
              placeholder="Search catalog..."
              type="text"
            />
          </div>

          <!-- Loading State -->
          <div v-if="appsStore.loading" class="loading-state">
            <div class="spinner"></div>
            <p>Loading catalog...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="appsStore.error" class="error-state">
            <p class="error-message">{{ appsStore.error }}</p>
            <Button variant="primary" @click="loadCatalog">RETRY</Button>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredApps.length === 0" class="empty-state">
            <h3>NO APPS FOUND</h3>
            <p class="text-secondary">Try adjusting your filters or search.</p>
          </div>

          <!-- App Grid -->
          <div v-else class="apps-grid">
            <Card
              v-for="app in filteredApps"
              :key="app.id"
              class="app-card hud-corners"
            >
              <div class="app-card-header">
                <div class="app-icon">{{ app.icon || '📦' }}</div>
                <div class="app-info">
                  <h3>{{ app.name }}</h3>
                  <p class="app-description">{{ app.description }}</p>
                </div>
              </div>

              <div class="app-meta">
                <span class="category-badge">{{ app.category }}</span>
                <span class="version-badge">v{{ app.version }}</span>
              </div>

              <div class="app-actions">
                <Button variant="primary" @click="goToAppDetail(app.id)">
                  INSTALL
                </Button>
              </div>
            </Card>
          </div>
        </div>

        <!-- DOCKER HUB TAB -->
        <div v-show="activeTab === 'dockerhub'" class="tab-content">
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

          <div v-else-if="searchError" class="error-state">
            <p class="error-message">{{ searchError }}</p>
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

    <!-- Installed Apps Modal -->
    <Modal v-model="showInstalledModal" title="INSTALLED APPS">
      <div v-if="installedApps.length === 0" class="empty-state">
        <p class="text-secondary">No apps installed yet.</p>
      </div>
      <div v-else class="installed-list">
        <div v-for="app in installedApps" :key="app.id" class="installed-item">
          <div class="installed-info">
            <h4>{{ app.appName }}</h4>
            <p class="text-secondary">Installed {{ formatDate(app.installedAt) }}</p>
            <span :class="['status-badge', app.status === 'running' ? 'status-online' : 'status-offline']">
              {{ app.status }}
            </span>
          </div>
          <div class="installed-actions">
            <Button variant="danger" size="sm" @click="confirmUninstall(app)">
              UNINSTALL
            </Button>
          </div>
        </div>
      </div>
    </Modal>

    <!-- Uninstall Confirmation Modal -->
    <Modal v-model="showUninstallModal" title="CONFIRM UNINSTALL">
      <p>Are you sure you want to uninstall <strong>{{ appToUninstall?.appName }}</strong>?</p>
      <p class="text-secondary mt-sm">This will remove the container and its data.</p>

      <template #footer>
        <Button variant="secondary" @click="showUninstallModal = false">
          CANCEL
        </Button>
        <Button variant="danger" @click="handleUninstall" :loading="uninstalling">
          {{ uninstalling ? 'UNINSTALLING...' : 'UNINSTALL' }}
        </Button>
      </template>
    </Modal>

    <!-- Toast -->
    <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppsStore } from '@/stores/apps'
import { useImagesStore } from '@/stores/images'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Modal from '@/components/ui/Modal.vue'
import Card from '@/components/ui/Card.vue'
import dayjs from 'dayjs'

const router = useRouter()
const appsStore = useAppsStore()
const imagesStore = useImagesStore()

// Tab state
const activeTab = ref('catalog')

// Catalog state
const catalogSearch = ref('')
const selectedCategory = ref('all')

const categories = [
  { value: 'all', label: 'ALL' },
  { value: 'monitoring', label: 'MONITORING' },
  { value: 'media', label: 'MEDIA' },
  { value: 'security', label: 'SECURITY' },
  { value: 'tools', label: 'TOOLS' },
  { value: 'productivity', label: 'PRODUCTIVITY' },
  { value: 'development', label: 'DEVELOPMENT' },
  { value: 'networking', label: 'NETWORKING' }
]

// Docker Hub state
const searchTerm = ref('')
const searching = ref(false)
const results = ref([])
const searchError = ref('')

// Pull modal state
const showPullModal = ref(false)
const pullImageName = ref('')
const pulling = ref(false)

// Installed apps state
const showInstalledModal = ref(false)
const installedApps = ref([])

// Uninstall state
const showUninstallModal = ref(false)
const appToUninstall = ref(null)
const uninstalling = ref(false)

// Toast
const toast = ref({ show: false, message: '', type: 'success' })

// Filtered apps for catalog
const filteredApps = computed(() => {
  let filtered = appsStore.apps

  // Category filter
  if (selectedCategory.value !== 'all') {
    filtered = filtered.filter(app => app.category === selectedCategory.value)
  }

  // Search filter
  if (catalogSearch.value.trim()) {
    const search = catalogSearch.value.toLowerCase()
    filtered = filtered.filter(app =>
      app.name.toLowerCase().includes(search) ||
      app.description.toLowerCase().includes(search)
    )
  }

  return filtered
})

async function loadCatalog() {
  try {
    await appsStore.fetchApps()
    await appsStore.fetchInstalledApps()
    installedApps.value = appsStore.installedApps
  } catch (err) {
    console.error('Failed to load catalog:', err)
  }
}

function goToAppDetail(appId) {
  router.push(`/apps/${appId}`)
}

// Docker Hub search
async function handleSearch() {
  if (!searchTerm.value.trim()) {
    searchError.value = 'Enter a search term.'
    results.value = []
    return
  }

  searchError.value = ''
  searching.value = true
  try {
    results.value = await imagesStore.searchImages(searchTerm.value.trim())
  } catch (err) {
    searchError.value = 'Search failed. Please try again.'
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

function confirmUninstall(app) {
  appToUninstall.value = app
  showUninstallModal.value = true
}

async function handleUninstall() {
  if (!appToUninstall.value) return

  uninstalling.value = true
  try {
    await appsStore.uninstallApp(appToUninstall.value.id)
    showToast(`Successfully uninstalled ${appToUninstall.value.appName}`, 'success')
    showUninstallModal.value = false
    await loadCatalog()
  } catch (err) {
    showToast('Failed to uninstall app', 'error')
  } finally {
    uninstalling.value = false
  }
}

function formatDate(dateString) {
  if (!dateString) return 'Unknown'
  return dayjs(dateString).format('MMM D, YYYY')
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  loadCatalog()
})
</script>

<style scoped>
.appstore-page {
  min-height: 100vh;
}

.appstore-main {
  padding-top: var(--space-2xl);
  padding-bottom: var(--space-3xl);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-xl);
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.installed-summary {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.summary-text {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--reach-cyan);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Tabs */
.tabs {
  display: flex;
  gap: var(--space-md);
  margin-bottom: var(--space-xl);
  border-bottom: 2px solid rgba(74, 85, 104, 0.3);
}

.tab {
  padding: var(--space-md) var(--space-lg);
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-base);
}

.tab:hover {
  color: var(--reach-amber);
}

.tab.active {
  color: var(--reach-amber);
  border-bottom-color: var(--reach-amber);
}

/* Filters */
.filters-section {
  display: flex;
  gap: var(--space-sm);
  flex-wrap: wrap;
  margin-bottom: var(--space-lg);
}

.filter-btn {
  padding: var(--space-xs) var(--space-md);
  background-color: transparent;
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-base);
}

.filter-btn:hover {
  border-color: var(--reach-amber);
  color: var(--reach-amber);
}

.filter-btn.active {
  background-color: var(--reach-amber);
  color: var(--reach-slate);
  border-color: var(--reach-amber);
}

/* Search */
.search-section {
  display: flex;
  gap: var(--space-md);
  align-items: center;
  margin-bottom: var(--space-2xl);
}

/* App Grid */
.apps-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-lg);
}

.app-card {
  padding: var(--space-lg);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  transition: all var(--transition-base);
}

.app-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.app-card-header {
  display: flex;
  gap: var(--space-md);
  align-items: flex-start;
}

.app-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.app-info h3 {
  margin-bottom: var(--space-xs);
  color: var(--reach-cyan);
}

.app-description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.app-meta {
  display: flex;
  gap: var(--space-sm);
}

.category-badge,
.version-badge {
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.65rem;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.category-badge {
  border: 1px solid rgba(34, 211, 238, 0.6);
  color: var(--reach-cyan);
}

.version-badge {
  border: 1px solid rgba(246, 166, 35, 0.6);
  color: var(--reach-amber);
}

.app-actions {
  margin-top: auto;
}

/* Docker Hub Results */
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

/* States */
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

/* Installed Apps Modal */
.installed-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.installed-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
}

.installed-info h4 {
  color: var(--reach-cyan);
  margin-bottom: var(--space-xs);
}

.status-badge {
  display: inline-block;
  margin-top: var(--space-xs);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.65rem;
  text-transform: uppercase;
}

.status-online {
  border: 1px solid var(--reach-cyan);
  color: var(--reach-cyan);
}

.status-offline {
  border: 1px solid var(--reach-orange);
  color: var(--reach-orange);
}

/* Pull Form */
.pull-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.text-sm {
  font-size: 0.875rem;
}

/* Toast */
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
  .page-header {
    flex-direction: column;
    gap: var(--space-md);
  }

  .search-section {
    flex-direction: column;
    align-items: stretch;
  }

  .apps-grid,
  .results-grid {
    grid-template-columns: 1fr;
  }
}
</style>
