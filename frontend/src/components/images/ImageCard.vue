<template>
  <Card class="image-card accent-bar" :show-corners="true">
    <div class="image-header">
      <h3 class="image-name">{{ getImageName() }}</h3>
      <Badge variant="online">{{ formatSize(image.Size) }}</Badge>
    </div>

    <div class="image-details">
      <div class="detail-row">
        <span class="label">ID</span>
        <span class="value font-mono">{{ image.Id.replace('sha256:', '').substring(0, 12) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">CREATED</span>
        <span class="value">{{ formatDate(image.Created) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">CONTAINERS</span>
        <span class="value">{{ image.Containers || 0 }} using this image</span>
      </div>
    </div>

    <div class="image-tags" v-if="image.RepoTags && image.RepoTags.length">
      <div class="tags-label">TAGS</div>
      <div class="tags-list">
        <span v-for="tag in image.RepoTags" :key="tag" class="tag">{{ tag }}</span>
      </div>
    </div>

    <div class="image-actions">
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('create', image)"
      >
        CREATE CONTAINER
      </Button>
      <Button
        variant="danger"
        size="sm"
        @click="$emit('remove', image.Id)"
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
  image: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['create', 'remove'])

function getImageName() {
  if (props.image.RepoTags && props.image.RepoTags.length > 0) {
    return props.image.RepoTags[0].split(':')[0]
  }
  return 'Unnamed Image'
}

function formatSize(bytes) {
  if (!bytes) return '0 MB'
  const mb = bytes / (1024 * 1024)
  if (mb >= 1000) {
    return `${(mb / 1024).toFixed(2)} GB`
  }
  return `${mb.toFixed(2)} MB`
}

function formatDate(timestamp) {
  return dayjs.unix(timestamp).format('YYYY-MM-DD HH:mm')
}
</script>

<style scoped>
.image-card {
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.image-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.image-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.image-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: var(--space-md);
}

.image-details {
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

.image-tags {
  margin-bottom: var(--space-md);
}

.tags-label {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
  margin-bottom: var(--space-sm);
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
}

.tag {
  padding: var(--space-xs) var(--space-sm);
  background-color: rgba(34, 211, 238, 0.1);
  border: 1px solid rgba(34, 211, 238, 0.3);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: 0.7rem;
  color: var(--reach-cyan);
}

.image-actions {
  display: flex;
  gap: var(--space-sm);
}

.image-actions .btn {
  flex: 1;
  font-size: 0.75rem;
  padding: var(--space-xs) var(--space-sm);
}
</style>
