<template>
  <div class="dashboard">
    <main class="dashboard-main">
      <div class="container">
        <div class="dashboard-header">
          <div>
            <h1>OPERATIONAL DASHBOARD</h1>
            <p class="subtitle">System Status and Container Management</p>
          </div>
          <div class="connection-status">
            <span class="status-label">METRICS STREAM</span>
            <div :class="['status-dot', wsConnected ? 'online' : 'warning', { 'status-pulse': wsConnected }]"></div>
            <span class="status-text">{{ wsConnected ? 'CONNECTED' : 'RECONNECTING' }}</span>
          </div>
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
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useSystemStore } from '@/stores/system'
import { useContainersStore } from '@/stores/containers'
import { useWebSocket } from '@/composables/useWebSocket'
import StatusPanel from '@/components/ui/StatusPanel.vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'

const systemStore = useSystemStore()
const containersStore = useContainersStore()

// WebSocket for real-time metrics
const { data: wsData, connected: wsConnected, connect: wsConnect, disconnect: wsDisconnect } = useWebSocket('/ws/metrics')

const cpuUsage = computed(() => {
  const usage = systemStore.metrics?.cpu?.usagePercent
  return typeof usage === 'number' ? `${usage.toFixed(1)}%` : '--'
})

const memoryUsage = computed(() => {
  const usage = systemStore.metrics?.memory?.usedPercent
  return typeof usage === 'number' ? `${usage.toFixed(1)}%` : '--'
})

const diskUsage = computed(() => {
  const usage = systemStore.metrics?.disk?.usedPercent
  return typeof usage === 'number' ? `${usage.toFixed(1)}%` : '--'
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

// Update system metrics when WebSocket message arrives
watch(wsData, (newData) => {
  if (newData && newData.type === 'metrics') {
    systemStore.metrics = newData.data
  }
})

// Fallback to polling if WebSocket disconnects
watch(wsConnected, (connected) => {
  if (!connected) {
    systemStore.startPolling(10000) // Poll every 10s as fallback
  } else {
    systemStore.stopPolling()
  }
})

onMounted(() => {
  wsConnect()
  containersStore.fetchContainers()
})

onUnmounted(() => {
  wsDisconnect()
  systemStore.stopPolling()
})
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
}

.dashboard-main {
  padding-top: var(--space-2xl);
  padding-bottom: var(--space-3xl);
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-2xl);
}

.dashboard-header h1 {
  margin-bottom: var(--space-sm);
}

.connection-status {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.status-label {
  color: var(--text-muted);
}

.status-text {
  color: var(--text-secondary);
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
