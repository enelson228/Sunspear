<template>
  <div id="app" class="reach-theme">
    <router-view />

    <!-- Container event toasts -->
    <div class="toast-container">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        :class="['toast', `toast-${toast.type}`]"
      >
        {{ toast.message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocket } from '@/composables/useWebSocket'

const authStore = useAuthStore()
const toasts = ref([])
let toastIdCounter = 0

// WebSocket for container events
const { data: eventData, connect: eventConnect, disconnect: eventDisconnect } = useWebSocket('/ws/events')

// Watch for authentication changes
watch(() => authStore.isAuthenticated, (isAuth) => {
  if (isAuth) {
    eventConnect()
  } else {
    eventDisconnect()
    toasts.value = []
  }
}, { immediate: true })

// Watch for container events
watch(eventData, (data) => {
  if (data && data.type === 'container') {
    showContainerEvent(data.data)
  }
})

function showContainerEvent(event) {
  const actionMap = {
    'start': { text: 'STARTED', type: 'success' },
    'stop': { text: 'STOPPED', type: 'warning' },
    'die': { text: 'DIED', type: 'warning' },
    'kill': { text: 'KILLED', type: 'warning' },
    'restart': { text: 'RESTARTED', type: 'success' }
  }

  const action = actionMap[event.action] || { text: event.action.toUpperCase(), type: 'info' }
  const containerName = event.containerName || event.containerId?.slice(0, 12) || 'Unknown'

  const toast = {
    id: toastIdCounter++,
    message: `${containerName} ${action.text}`,
    type: action.type
  }

  // Add toast
  toasts.value.push(toast)

  // Keep max 3 toasts
  if (toasts.value.length > 3) {
    toasts.value.shift()
  }

  // Auto-dismiss after 4 seconds
  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== toast.id)
  }, 4000)
}
</script>

<style>
#app {
  min-height: 100vh;
  background-color: var(--reach-slate);
  color: var(--text-primary);
}

.toast-container {
  position: fixed;
  bottom: var(--space-xl);
  right: var(--space-xl);
  z-index: var(--z-toast);
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
  pointer-events: none;
}

.toast {
  padding: var(--space-md) var(--space-lg);
  background-color: var(--reach-steel);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  animation: slideIn 0.3s ease-out;
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  min-width: 250px;
}

.toast-success {
  border: 1px solid var(--reach-cyan);
  color: var(--reach-cyan);
}

.toast-warning {
  border: 1px solid var(--reach-orange);
  color: var(--reach-orange);
}

.toast-info {
  border: 1px solid var(--reach-amber);
  color: var(--reach-amber);
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>
