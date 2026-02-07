<template>
  <div class="appdetail-page">
    <main class="appdetail-main">
      <div class="container">
        <Button variant="secondary" @click="$router.push('/apps')" class="mb-lg">
          ← BACK TO APP STORE
        </Button>

        <!-- Loading State -->
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Loading app details...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="loadAppDetails">RETRY</Button>
        </div>

        <!-- App Details & Installation Form -->
        <div v-else-if="app" class="detail-content">
          <!-- App Header -->
          <Card class="app-header-card hud-corners">
            <div class="app-header-content">
              <div class="app-icon-large">{{ app.icon || '📦' }}</div>
              <div class="app-header-info">
                <h1>{{ app.name }}</h1>
                <p class="app-description">{{ app.description }}</p>
                <div class="app-badges">
                  <span class="category-badge">{{ app.category }}</span>
                  <span class="version-badge">v{{ app.version }}</span>
                </div>
              </div>
            </div>
          </Card>

          <!-- Installation Form -->
          <Card class="install-form-card hud-corners">
            <h2 class="section-title">INSTALLATION CONFIGURATION</h2>

            <form @submit.prevent="handleInstall" class="install-form">
              <!-- Container Name -->
              <div class="form-group">
                <Input
                  v-model="config.name"
                  label="CONTAINER NAME"
                  placeholder="my-app"
                  type="text"
                  :disabled="installing"
                />
              </div>

              <!-- Port Mappings -->
              <div class="form-section" v-if="app.ports && Object.keys(app.ports).length > 0">
                <h3 class="form-section-title">PORT MAPPINGS</h3>
                <p class="form-section-description">Configure which ports to expose on the host.</p>
                <div class="port-mappings">
                  <div
                    v-for="(containerPort, key) in app.ports"
                    :key="key"
                    class="port-mapping-row"
                  >
                    <span class="port-label">{{ key }}</span>
                    <Input
                      v-model="config.ports[key]"
                      placeholder="Host port"
                      type="number"
                      :disabled="installing"
                    />
                    <span class="port-arrow">→</span>
                    <span class="port-value">{{ containerPort }}</span>
                  </div>
                </div>
              </div>

              <!-- Volume Paths -->
              <div class="form-section" v-if="app.volumes && app.volumes.length > 0">
                <h3 class="form-section-title">VOLUME PATHS</h3>
                <p class="form-section-description">Specify where to store persistent data on the host.</p>
                <div class="volume-mappings">
                  <div
                    v-for="volume in app.volumes"
                    :key="volume.container"
                    class="volume-mapping-row"
                  >
                    <Input
                      v-model="config.volumes[volume.container]"
                      :label="volume.container"
                      :placeholder="volume.host || '/path/on/host'"
                      type="text"
                      :disabled="installing"
                    />
                  </div>
                </div>
              </div>

              <!-- Required Environment Variables -->
              <div class="form-section" v-if="app.envVars?.required && app.envVars.required.length > 0">
                <h3 class="form-section-title">REQUIRED ENVIRONMENT VARIABLES</h3>
                <p class="form-section-description">These environment variables are required for the app to function.</p>
                <div class="env-vars">
                  <div
                    v-for="envVar in app.envVars.required"
                    :key="envVar.name"
                    class="env-var-row"
                  >
                    <Input
                      v-model="config.env[envVar.name]"
                      :label="envVar.name"
                      :placeholder="envVar.description || 'Enter value...'"
                      type="text"
                      :disabled="installing"
                    />
                    <p v-if="envVar.description" class="env-description">{{ envVar.description }}</p>
                  </div>
                </div>
              </div>

              <!-- Optional Environment Variables -->
              <div class="form-section" v-if="app.envVars?.optional && app.envVars.optional.length > 0">
                <h3 class="form-section-title">OPTIONAL ENVIRONMENT VARIABLES</h3>
                <p class="form-section-description">These environment variables are optional and have defaults.</p>
                <div class="env-vars">
                  <div
                    v-for="envVar in app.envVars.optional"
                    :key="envVar.name"
                    class="env-var-row"
                  >
                    <Input
                      v-model="config.env[envVar.name]"
                      :label="envVar.name"
                      :placeholder="envVar.default || envVar.description || 'Enter value...'"
                      type="text"
                      :disabled="installing"
                    />
                    <p v-if="envVar.description" class="env-description">{{ envVar.description }}</p>
                  </div>
                </div>
              </div>

              <!-- Install Button -->
              <div class="form-actions">
                <Button
                  type="submit"
                  variant="primary"
                  :loading="installing"
                  :disabled="!isFormValid"
                >
                  {{ installing ? 'INSTALLING...' : 'INSTALL APPLICATION' }}
                </Button>
              </div>
            </form>
          </Card>

          <!-- Installation Status -->
          <Card v-if="installStatus" class="status-card hud-corners" :class="`status-${installStatus.type}`">
            <h3 class="status-title">{{ installStatus.title }}</h3>
            <p class="status-message">{{ installStatus.message }}</p>
            <Button v-if="installStatus.type === 'success'" variant="primary" @click="goToContainer">
              VIEW CONTAINER
            </Button>
          </Card>
        </div>
      </div>
    </main>

    <!-- Toast -->
    <div v-if="toast.show" class="toast" :class="`toast-${toast.type}`">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppsStore } from '@/stores/apps'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Card from '@/components/ui/Card.vue'

const router = useRouter()
const route = useRoute()
const appsStore = useAppsStore()

const app = ref(null)
const loading = ref(false)
const error = ref(null)
const installing = ref(false)
const installStatus = ref(null)
const installedContainerId = ref(null)

const config = ref({
  name: '',
  ports: {},
  volumes: {},
  env: {}
})

const toast = ref({ show: false, message: '', type: 'success' })

const appId = computed(() => route.params.id)

const isFormValid = computed(() => {
  // Check container name
  if (!config.value.name || !config.value.name.trim()) {
    return false
  }

  // Check required env vars
  if (app.value?.envVars?.required) {
    for (const envVar of app.value.envVars.required) {
      if (!config.value.env[envVar.name] || !config.value.env[envVar.name].trim()) {
        return false
      }
    }
  }

  return true
})

async function loadAppDetails() {
  loading.value = true
  error.value = null
  try {
    app.value = await appsStore.getApp(appId.value)
    initializeConfig()
  } catch (err) {
    error.value = err.message || 'Failed to load app details'
  } finally {
    loading.value = false
  }
}

function initializeConfig() {
  // Pre-fill container name with app ID
  config.value.name = app.value.id

  // Pre-fill port mappings
  if (app.value.ports) {
    for (const [key, containerPort] of Object.entries(app.value.ports)) {
      config.value.ports[key] = containerPort // Default to same port
    }
  }

  // Pre-fill volume paths with suggestions
  if (app.value.volumes) {
    for (const volume of app.value.volumes) {
      config.value.volumes[volume.container] = volume.host || ''
    }
  }

  // Initialize env vars
  if (app.value.envVars?.optional) {
    for (const envVar of app.value.envVars.optional) {
      if (envVar.default) {
        config.value.env[envVar.name] = envVar.default
      }
    }
  }
}

async function handleInstall() {
  if (!isFormValid.value) {
    showToast('Please fill in all required fields', 'error')
    return
  }

  installing.value = true
  installStatus.value = null

  try {
    // Format environment variables as array
    const envArray = Object.entries(config.value.env)
      .filter(([key, value]) => value && value.trim())
      .map(([key, value]) => ({ name: key, value: value.trim() }))

    // Convert port values to strings for the backend
    const portsAsStrings = {}
    for (const [key, value] of Object.entries(config.value.ports)) {
      portsAsStrings[key] = String(value)
    }

    const installPayload = {
      name: config.value.name.trim(),
      env: envArray,
      ports: portsAsStrings,
      volumes: config.value.volumes
    }

    const result = await appsStore.installApp(appId.value, installPayload)

    installStatus.value = {
      type: 'success',
      title: 'Installation Successful',
      message: `${app.value.name} has been installed successfully. The container is now running.`
    }

    installedContainerId.value = result.containerId

    showToast('Application installed successfully', 'success')
  } catch (err) {
    installStatus.value = {
      type: 'error',
      title: 'Installation Failed',
      message: err.message || 'Failed to install application. Please check the configuration and try again.'
    }

    showToast('Installation failed', 'error')
  } finally {
    installing.value = false
  }
}

function goToContainer() {
  if (installedContainerId.value) {
    router.push(`/containers/${installedContainerId.value}`)
  } else {
    router.push('/containers')
  }
}

function showToast(message, type = 'success') {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  loadAppDetails()
})
</script>

<style scoped>
.appdetail-page {
  min-height: 100vh;
}

.appdetail-main {
  padding-top: var(--space-2xl);
  padding-bottom: var(--space-3xl);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

/* App Header */
.app-header-card {
  padding: var(--space-xl);
}

.app-header-content {
  display: flex;
  gap: var(--space-lg);
  align-items: flex-start;
}

.app-icon-large {
  font-size: 4rem;
  flex-shrink: 0;
}

.app-header-info h1 {
  margin-bottom: var(--space-sm);
  color: var(--reach-cyan);
}

.app-description {
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin-bottom: var(--space-md);
}

.app-badges {
  display: flex;
  gap: var(--space-sm);
}

.category-badge,
.version-badge {
  padding: 4px 12px;
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.75rem;
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

/* Installation Form */
.install-form-card {
  padding: var(--space-xl);
}

.section-title {
  font-size: 1.25rem;
  margin-bottom: var(--space-lg);
  color: var(--reach-amber);
}

.install-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

.form-group {
  width: 100%;
}

.form-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.form-section-title {
  font-size: 1rem;
  color: var(--reach-cyan);
  margin-bottom: var(--space-xs);
}

.form-section-description {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin-bottom: var(--space-sm);
}

/* Port Mappings */
.port-mappings {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.port-mapping-row {
  display: grid;
  grid-template-columns: 150px 1fr auto 100px;
  gap: var(--space-md);
  align-items: center;
}

.port-label {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
}

.port-arrow {
  color: var(--reach-amber);
  font-size: 1.25rem;
}

.port-value {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--reach-cyan);
  text-align: center;
}

/* Volume Mappings */
.volume-mappings {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.volume-mapping-row {
  width: 100%;
}

/* Environment Variables */
.env-vars {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.env-var-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.env-description {
  font-size: 0.75rem;
  color: var(--text-muted);
  font-family: var(--font-mono);
}

/* Form Actions */
.form-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: var(--space-lg);
  border-top: 1px solid rgba(74, 85, 104, 0.3);
}

/* Status Card */
.status-card {
  padding: var(--space-xl);
  text-align: center;
}

.status-success {
  border: 2px solid var(--reach-cyan);
}

.status-error {
  border: 2px solid var(--reach-orange);
}

.status-title {
  font-size: 1.5rem;
  margin-bottom: var(--space-md);
}

.status-success .status-title {
  color: var(--reach-cyan);
}

.status-error .status-title {
  color: var(--reach-orange);
}

.status-message {
  color: var(--text-secondary);
  margin-bottom: var(--space-lg);
  line-height: 1.6;
}

/* States */
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
  .app-header-content {
    flex-direction: column;
  }

  .port-mapping-row {
    grid-template-columns: 1fr;
    gap: var(--space-sm);
  }

  .port-arrow {
    display: none;
  }
}
</style>
