<template>
  <div class="dashboard">
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

    <main class="dashboard-main">
      <div class="container">
        <div class="dashboard-header">
          <h1>OPERATIONAL DASHBOARD</h1>
          <p class="subtitle">System Status and Container Management</p>
        </div>

        <div class="metrics-grid">
          <StatusPanel
            title="CPU USAGE"
            :value="cpuUsage"
            label="Processor Load"
            :status="cpuStatus"
          />
          <StatusPanel
            title="MEMORY"
            :value="memoryUsage"
            label="RAM Utilization"
            :status="memoryStatus"
          />
          <StatusPanel
            title="DISK"
            :value="diskUsage"
            label="Storage Used"
            :status="diskStatus"
          />
          <StatusPanel
            title="CONTAINERS"
            :value="containerCount"
            label="Active Containers"
            status="online"
          />
        </div>

        <div class="quick-actions mt-xl">
          <h2 class="mb-lg">QUICK ACTIONS</h2>
          <div class="actions-grid">
            <Card class="action-card accent-bar">
              <h3>View Containers</h3>
              <p class="text-secondary">Manage running containers</p>
              <Button variant="primary" class="mt-md" @click="$router.push('/containers')">
                OPEN
              </Button>
            </Card>
            <Card class="action-card accent-bar">
              <h3>Browse App Store</h3>
              <p class="text-secondary">Install new applications</p>
              <Button variant="primary" class="mt-md" @click="$router.push('/apps')">
                OPEN
              </Button>
            </Card>
            <Card class="action-card accent-bar">
              <h3>System Monitor</h3>
              <p class="text-secondary">Detailed system metrics</p>
              <Button variant="primary" class="mt-md" @click="$router.push('/system')">
                OPEN
              </Button>
            </Card>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import { useContainersStore } from '@/stores/containers'
import StatusPanel from '@/components/ui/StatusPanel.vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const authStore = useAuthStore()
const systemStore = useSystemStore()
const containersStore = useContainersStore()

const cpuUsage = computed(() => {
  return systemStore.metrics?.cpu.usagePercent.toFixed(1) + '%' || '--'
})

const memoryUsage = computed(() => {
  return systemStore.metrics?.memory.usedPercent.toFixed(1) + '%' || '--'
})

const diskUsage = computed(() => {
  return systemStore.metrics?.disk.usedPercent.toFixed(1) + '%' || '--'
})

const containerCount = computed(() => {
  return containersStore.containers.filter(c => c.State === 'running').length || 0
})

const cpuStatus = computed(() => {
  const usage = systemStore.metrics?.cpu.usagePercent || 0
  if (usage > 80) return 'warning'
  return 'online'
})

const memoryStatus = computed(() => {
  const usage = systemStore.metrics?.memory.usedPercent || 0
  if (usage > 85) return 'warning'
  return 'online'
})

const diskStatus = computed(() => {
  const usage = systemStore.metrics?.disk.usedPercent || 0
  if (usage > 90) return 'warning'
  return 'online'
})

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  systemStore.startPolling()
  containersStore.fetchContainers()
})

onUnmounted(() => {
  systemStore.stopPolling()
})
</script>

<style scoped>
.dashboard {
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

.dashboard-main {
  padding-top: 100px;
  padding-bottom: var(--space-3xl);
}

.dashboard-header {
  margin-bottom: var(--space-2xl);
}

.dashboard-header h1 {
  margin-bottom: var(--space-sm);
}

.subtitle {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-lg);
}

.quick-actions h2 {
  font-size: 1.5rem;
  margin-bottom: var(--space-lg);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--space-lg);
}

.action-card {
  padding: var(--space-xl);
}

.action-card h3 {
  margin-bottom: var(--space-sm);
  color: var(--reach-amber);
}

.action-card p {
  margin-bottom: var(--space-md);
}
</style>
