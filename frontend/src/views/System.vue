<template>
  <div class="system-page">
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

    <main class="system-main">
      <div class="container">
        <div class="page-header">
          <div>
            <h1>SYSTEM MONITORING</h1>
            <p class="subtitle">Real-time performance metrics and diagnostics</p>
          </div>
          <div class="header-actions">
            <Badge variant="online" :pulse="true">LIVE</Badge>
          </div>
        </div>

        <div v-if="loading && !metrics" class="loading-state">
          <div class="spinner"></div>
          <p>Loading system metrics...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <p class="error-message">{{ error }}</p>
          <Button variant="primary" @click="systemStore.fetchMetrics">RETRY</Button>
        </div>

        <div v-else-if="metrics" class="metrics-content">
          <div class="gauges-grid">
            <MetricGauge
              title="CPU USAGE"
              :value="metrics.cpu.usagePercent"
              label="Processor Load"
            />
            <MetricGauge
              title="MEMORY"
              :value="metrics.memory.usedPercent"
              label="RAM Utilization"
            />
            <MetricGauge
              title="DISK USAGE"
              :value="metrics.disk.usedPercent"
              label="Storage Used"
            />
          </div>

          <div class="details-grid">
            <Card class="detail-card">
              <h3 class="section-title">CPU DETAILS</h3>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">CORES</span>
                  <span class="detail-value font-mono">{{ metrics.cpu.cores }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">USAGE</span>
                  <span class="detail-value font-mono">{{ metrics.cpu.usagePercent.toFixed(2) }}%</span>
                </div>
              </div>
            </Card>

            <Card class="detail-card">
              <h3 class="section-title">MEMORY DETAILS</h3>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">TOTAL</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.memory.total) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">USED</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.memory.used) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">AVAILABLE</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.memory.available) }}</span>
                </div>
              </div>
            </Card>

            <Card class="detail-card">
              <h3 class="section-title">DISK DETAILS</h3>
              <div class="detail-rows">
                <div class="detail-row">
                  <span class="detail-label">TOTAL</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.disk.total) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">USED</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.disk.used) }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">FREE</span>
                  <span class="detail-value font-mono">{{ formatBytes(metrics.disk.free) }}</span>
                </div>
              </div>
            </Card>
          </div>

          <div class="charts-grid">
            <MetricChart
              title="CPU USAGE HISTORY"
              subtitle="Last 60 seconds"
              :data="cpuHistory"
              :labels="timeLabels"
              color="#22d3ee"
            />
            <MetricChart
              title="MEMORY USAGE HISTORY"
              subtitle="Last 60 seconds"
              :data="memoryHistory"
              :labels="timeLabels"
              color="#f6a623"
            />
          </div>

          <Card class="network-card">
            <h3 class="section-title">NETWORK STATISTICS</h3>
            <div class="network-grid">
              <div class="network-stat">
                <span class="network-label">BYTES SENT</span>
                <span class="network-value font-mono">{{ formatBytes(metrics.network.bytesSent) }}</span>
              </div>
              <div class="network-stat">
                <span class="network-label">BYTES RECEIVED</span>
                <span class="network-value font-mono">{{ formatBytes(metrics.network.bytesRecv) }}</span>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSystemStore } from '@/stores/system'
import MetricGauge from '@/components/system/MetricGauge.vue'
import MetricChart from '@/components/system/MetricChart.vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import dayjs from 'dayjs'

const router = useRouter()
const authStore = useAuthStore()
const systemStore = useSystemStore()

const cpuHistory = ref([])
const memoryHistory = ref([])
const timeLabels = ref([])
const maxDataPoints = 60

const metrics = computed(() => systemStore.metrics)
const loading = computed(() => systemStore.loading)
const error = computed(() => systemStore.error)

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

function updateHistory() {
  if (!metrics.value) return

  cpuHistory.value.push(metrics.value.cpu.usagePercent)
  memoryHistory.value.push(metrics.value.memory.usedPercent)
  timeLabels.value.push(dayjs().format('HH:mm:ss'))

  if (cpuHistory.value.length > maxDataPoints) {
    cpuHistory.value.shift()
    memoryHistory.value.shift()
    timeLabels.value.shift()
  }
}

onMounted(() => {
  systemStore.startPolling(5000)

  const historyInterval = setInterval(updateHistory, 5000)

  onUnmounted(() => {
    systemStore.stopPolling()
    clearInterval(historyInterval)
  })
})
</script>

<style scoped>
.system-page {
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

.system-main {
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

.metrics-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

.gauges-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--space-lg);
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: var(--space-lg);
}

.detail-card {
  padding: var(--space-lg);
}

.section-title {
  font-size: 1rem;
  margin-bottom: var(--space-md);
  color: var(--reach-amber);
}

.detail-rows {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid rgba(74, 85, 104, 0.2);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.detail-value {
  color: var(--text-primary);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: var(--space-lg);
}

.network-card {
  padding: var(--space-lg);
}

.network-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-lg);
}

.network-stat {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-sm);
}

.network-label {
  font-family: var(--font-mono);
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.network-value {
  font-size: 1.25rem;
  color: var(--reach-cyan);
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

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: var(--space-md);
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }
}
</style>
