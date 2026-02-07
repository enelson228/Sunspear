<template>
  <Card class="project-card accent-bar" :show-corners="true">
    <div class="project-header">
      <h3 class="project-name">{{ project.name }}</h3>
      <Badge :variant="getStatusVariant()">{{ project.status || 'UNKNOWN' }}</Badge>
    </div>

    <div class="project-details">
      <div class="detail-row" v-if="project.description">
        <span class="label">DESCRIPTION</span>
        <span class="value">{{ project.description }}</span>
      </div>
      <div class="detail-row">
        <span class="label">CONTAINERS</span>
        <span class="value">{{ project.container_count || 0 }} containers</span>
      </div>
      <div class="detail-row">
        <span class="label">CREATED</span>
        <span class="value">{{ formatDate(project.created_at) }}</span>
      </div>
    </div>

    <div class="project-actions">
      <Button
        v-if="project.status !== 'running'"
        variant="secondary"
        size="sm"
        @click="$emit('start', project.id)"
        :disabled="loading"
      >
        START
      </Button>
      <Button
        v-if="project.status === 'running'"
        variant="secondary"
        size="sm"
        @click="$emit('stop', project.id)"
        :disabled="loading"
      >
        STOP
      </Button>
      <Button
        v-if="project.status === 'running'"
        variant="secondary"
        size="sm"
        @click="$emit('restart', project.id)"
        :disabled="loading"
      >
        RESTART
      </Button>
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('details', project.id)"
      >
        DETAILS
      </Button>
      <Button
        variant="danger"
        size="sm"
        @click="$emit('delete', project.id)"
        :disabled="loading"
      >
        DELETE
      </Button>
    </div>
  </Card>
</template>

<script setup>
import Card from '@/components/ui/Card.vue'
import Badge from '@/components/ui/Badge.vue'
import Button from '@/components/ui/Button.vue'
import dayjs from 'dayjs'

const props = defineProps({
  project: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['start', 'stop', 'restart', 'details', 'delete'])

function getStatusVariant() {
  const status = props.project.status?.toLowerCase()
  if (status === 'running') return 'online'
  if (status === 'stopped') return 'warning'
  if (status === 'error') return 'offline'
  return 'warning'
}

function formatDate(timestamp) {
  if (!timestamp) return 'N/A'
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm')
}
</script>

<style scoped>
.project-card {
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.project-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.project-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.project-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: var(--space-md);
}

.project-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 0.875rem;
}

.detail-row .label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  min-width: 120px;
}

.detail-row .value {
  color: var(--text-primary);
  text-align: right;
  flex: 1;
}

.project-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.project-actions .btn {
  font-size: 0.75rem;
  padding: var(--space-xs) var(--space-sm);
}
</style>
