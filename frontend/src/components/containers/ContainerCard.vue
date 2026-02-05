<template>
  <Card class="container-card accent-bar" :show-corners="true">
    <div class="container-card-header">
      <div class="container-info">
        <h3 class="container-name">{{ container.Names[0]?.replace('/', '') || 'Unknown' }}</h3>
        <Badge :variant="statusVariant" :pulse="container.State === 'running'">
          {{ container.State }}
        </Badge>
      </div>
      <div class="container-id font-mono text-muted">{{ container.Id.substring(0, 12) }}</div>
    </div>

    <div class="container-details">
      <div class="detail-row">
        <span class="label">IMAGE</span>
        <span class="value font-mono">{{ container.Image }}</span>
      </div>
      <div class="detail-row">
        <span class="label">PORTS</span>
        <span class="value font-mono">{{ formatPorts(container.Ports) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">STATUS</span>
        <span class="value">{{ container.Status }}</span>
      </div>
    </div>

    <div class="container-actions">
      <Button
        v-if="container.State !== 'running'"
        variant="primary"
        size="sm"
        @click="$emit('start', container.Id)"
        :disabled="loading"
      >
        START
      </Button>
      <Button
        v-if="container.State === 'running'"
        variant="secondary"
        size="sm"
        @click="$emit('stop', container.Id)"
        :disabled="loading"
      >
        STOP
      </Button>
      <Button
        v-if="container.State === 'running'"
        variant="secondary"
        size="sm"
        @click="$emit('restart', container.Id)"
        :disabled="loading"
      >
        RESTART
      </Button>
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('view', container.Id)"
      >
        DETAILS
      </Button>
      <Button
        variant="danger"
        size="sm"
        @click="$emit('remove', container.Id)"
        :disabled="loading"
      >
        REMOVE
      </Button>
    </div>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'

const props = defineProps({
  container: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['start', 'stop', 'restart', 'view', 'remove'])

const statusVariant = computed(() => {
  switch (props.container.State) {
    case 'running':
      return 'online'
    case 'exited':
      return 'offline'
    case 'paused':
      return 'warning'
    default:
      return 'offline'
  }
})

function formatPorts(ports) {
  if (!ports || ports.length === 0) return 'None'
  return ports
    .map(p => p.PublicPort ? `${p.PublicPort}:${p.PrivatePort}` : p.PrivatePort)
    .join(', ')
}
</script>

<style scoped>
.container-card {
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.container-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.container-card-header {
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.container-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-sm);
}

.container-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
}

.container-id {
  font-size: 0.75rem;
}

.container-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  margin-bottom: var(--space-lg);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.detail-row .label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  min-width: 80px;
}

.detail-row .value {
  color: var(--text-primary);
  text-align: right;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.container-actions {
  display: flex;
  gap: var(--space-sm);
  flex-wrap: wrap;
}

.container-actions .btn {
  flex: 1;
  min-width: 80px;
  font-size: 0.75rem;
  padding: var(--space-xs) var(--space-sm);
}
</style>
