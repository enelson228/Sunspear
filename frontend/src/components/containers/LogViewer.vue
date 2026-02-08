<template>
  <div class="log-viewer hud-corners">
    <div class="log-header">
      <div class="log-title">
        <span :class="['status-dot', streaming ? 'online' : 'offline', { 'status-pulse': streaming }]"></span>
        <span class="font-mono">{{ streaming ? 'STREAMING LOGS' : 'CONTAINER LOGS' }}</span>
      </div>
      <div class="log-controls">
        <Button
          variant="secondary"
          size="sm"
          @click="toggleStreaming"
          :class="{ 'btn-active': streaming }"
        >
          {{ streaming ? 'STOP STREAM' : 'STREAM' }}
        </Button>
        <Button variant="secondary" size="sm" @click="$emit('refresh')" :disabled="streaming">
          REFRESH
        </Button>
        <Button variant="secondary" size="sm" @click="$emit('download')">
          DOWNLOAD
        </Button>
      </div>
    </div>
    <div class="log-content" ref="logContainer">
      <div v-if="loading" class="log-loading">
        <div class="spinner"></div>
        <span>Loading logs...</span>
      </div>
      <pre v-else-if="displayLogs" class="log-text">{{ displayLogs }}</pre>
      <div v-else class="log-empty">
        <span>No logs available</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick, computed } from 'vue'
import { useWebSocket } from '@/composables/useWebSocket'
import Button from '@/components/ui/Button.vue'

const props = defineProps({
  logs: String,
  loading: Boolean,
  containerId: String,
  autoScroll: {
    type: Boolean,
    default: true
  }
})

defineEmits(['refresh', 'download'])

const logContainer = ref(null)
const streaming = ref(false)
const streamedLogs = ref('')

// WebSocket for log streaming
const { data: wsData, connected: wsConnected, connect: wsConnect, disconnect: wsDisconnect } = useWebSocket(`/ws/logs/${props.containerId}`)

// Display either streamed logs or regular logs
const displayLogs = computed(() => {
  return streaming.value ? streamedLogs.value : props.logs
})

function toggleStreaming() {
  if (streaming.value) {
    stopStreaming()
  } else {
    startStreaming()
  }
}

function startStreaming() {
  streamedLogs.value = props.logs || '' // Start with existing logs
  streaming.value = true
  wsConnect()
}

function stopStreaming() {
  streaming.value = false
  wsDisconnect()
}

// Watch for incoming log messages
watch(wsData, (data) => {
  if (data && data.type === 'log' && streaming.value) {
    streamedLogs.value += data.data
    scrollToBottom()
  }
})

// Watch for regular logs updates
watch(() => props.logs, async () => {
  if (!streaming.value && props.autoScroll) {
    await nextTick()
    scrollToBottom()
  }
})

async function scrollToBottom() {
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}
</script>

<style scoped>
.log-viewer {
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-lg);
  background-color: var(--reach-steel);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.log-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.log-controls {
  display: flex;
  gap: var(--space-sm);
}

.log-content {
  height: 400px;
  overflow-y: auto;
  padding: var(--space-md);
  font-family: var(--font-mono);
  font-size: 0.75rem;
  line-height: 1.5;
}

.log-text {
  margin: 0;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-wrap: break-word;
}

.log-loading,
.log-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: var(--space-sm);
  color: var(--text-secondary);
}

/* Custom scrollbar for log viewer */
.log-content::-webkit-scrollbar {
  width: 10px;
}

.log-content::-webkit-scrollbar-track {
  background: var(--reach-slate);
}

.log-content::-webkit-scrollbar-thumb {
  background: var(--reach-titanium);
  border-radius: 5px;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: var(--reach-amber);
}

.btn-active {
  background-color: var(--reach-titanium);
  border-color: var(--reach-cyan);
  color: var(--reach-cyan);
}
</style>
