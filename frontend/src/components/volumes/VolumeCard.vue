<template>
  <Card class="volume-card accent-bar" :show-corners="true">
    <div class="volume-header">
      <h3 class="volume-name">{{ volume.Name }}</h3>
      <Badge variant="online">{{ volume.Driver }}</Badge>
    </div>

    <div class="volume-details">
      <div class="detail-row">
        <span class="label">MOUNTPOINT</span>
        <span class="value font-mono">{{ formatMountpoint(volume.Mountpoint) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">SCOPE</span>
        <span class="value">{{ volume.Scope || 'local' }}</span>
      </div>
      <div class="detail-row">
        <span class="label">CREATED</span>
        <span class="value">{{ formatDate(volume.CreatedAt) }}</span>
      </div>
    </div>

    <div class="volume-labels" v-if="volume.Labels && Object.keys(volume.Labels).length">
      <div class="labels-title">LABELS</div>
      <div class="labels-list">
        <span v-for="(value, key) in volume.Labels" :key="key" class="label-item">
          {{ key }}: {{ value }}
        </span>
      </div>
    </div>

    <div class="volume-actions">
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('inspect', volume.Name)"
      >
        INSPECT
      </Button>
      <Button
        variant="danger"
        size="sm"
        @click="$emit('remove', volume.Name)"
        :disabled="loading"
      >
        REMOVE
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
  volume: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['inspect', 'remove'])

function formatMountpoint(path) {
  if (!path) return 'N/A'
  const parts = path.split('/')
  if (parts.length > 3) {
    return '.../' + parts.slice(-2).join('/')
  }
  return path
}

function formatDate(timestamp) {
  if (!timestamp) return 'N/A'
  return dayjs(timestamp).format('YYYY-MM-DD HH:mm')
}
</script>

<style scoped>
.volume-card {
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.volume-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.volume-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.volume-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: var(--space-md);
}

.volume-details {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
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
  min-width: 100px;
}

.detail-row .value {
  color: var(--text-primary);
  text-align: right;
  flex: 1;
}

.volume-labels {
  margin-bottom: var(--space-md);
}

.labels-title {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  margin-bottom: var(--space-sm);
}

.labels-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.label-item {
  padding: var(--space-xs) var(--space-sm);
  background-color: rgba(34, 211, 238, 0.1);
  border: 1px solid rgba(34, 211, 238, 0.3);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.7rem;
  color: var(--reach-cyan);
}

.volume-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.volume-actions .btn {
  font-size: 0.75rem;
  padding: var(--space-xs) var(--space-sm);
}
</style>
