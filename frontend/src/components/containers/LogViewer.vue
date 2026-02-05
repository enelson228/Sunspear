<template>
  <div class="log-viewer hud-corners">
    <div class="log-header">
      <div class="log-title">
        <span class="status-dot online status-pulse"></span>
        <span class="font-mono">CONTAINER LOGS</span>
      </div>
      <div class="log-controls">
        <Button variant="secondary" size="sm" @click="$emit('refresh')">
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
      <pre v-else-if="logs" class="log-text">{{ logs }}</pre>
      <div v-else class="log-empty">
        <span>No logs available</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import Button from '@/components/ui/Button.vue'

const props = defineProps({
  logs: String,
  loading: Boolean,
  autoScroll: {
    type: Boolean,
    default: true
  }
})

defineEmits(['refresh', 'download'])

const logContainer = ref(null)

watch(() => props.logs, async () => {
  if (props.autoScroll) {
    await nextTick()
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  }
})
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
</style>
