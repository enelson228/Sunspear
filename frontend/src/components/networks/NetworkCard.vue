<template>
  <Card class="network-card accent-bar" :show-corners="true">
    <div class="network-header">
      <h3 class="network-name">{{ network.Name }}</h3>
      <Badge variant="online">{{ network.Driver }}</Badge>
    </div>

    <div class="network-details">
      <div class="detail-row">
        <span class="label">ID</span>
        <span class="value font-mono">{{ network.Id.substring(0, 12) }}</span>
      </div>
      <div class="detail-row">
        <span class="label">SCOPE</span>
        <span class="value">{{ network.Scope || 'local' }}</span>
      </div>
      <div class="detail-row" v-if="getSubnet()">
        <span class="label">SUBNET</span>
        <span class="value font-mono">{{ getSubnet() }}</span>
      </div>
      <div class="detail-row" v-if="getGateway()">
        <span class="label">GATEWAY</span>
        <span class="value font-mono">{{ getGateway() }}</span>
      </div>
      <div class="detail-row">
        <span class="label">CONTAINERS</span>
        <span class="value">{{ getContainerCount() }} connected</span>
      </div>
    </div>

    <div class="network-actions">
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('inspect', network.Id)"
      >
        INSPECT
      </Button>
      <Button
        variant="danger"
        size="sm"
        @click="$emit('remove', network.Id)"
        :disabled="loading || isBuiltIn()"
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

const props = defineProps({
  network: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['inspect', 'remove'])

function getSubnet() {
  if (!props.network.IPAM?.Config?.length) return null
  return props.network.IPAM.Config[0]?.Subnet || null
}

function getGateway() {
  if (!props.network.IPAM?.Config?.length) return null
  return props.network.IPAM.Config[0]?.Gateway || null
}

function getContainerCount() {
  if (!props.network.Containers) return 0
  return Object.keys(props.network.Containers).length
}

function isBuiltIn() {
  const builtInNetworks = ['bridge', 'host', 'none']
  return builtInNetworks.includes(props.network.Name)
}
</script>

<style scoped>
.network-card {
  padding: var(--space-lg);
  transition: all var(--transition-base);
}

.network-card:hover {
  transform: translateY(-4px);
  border-color: var(--reach-amber);
}

.network-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
  padding-bottom: var(--space-md);
  border-bottom: 1px solid rgba(74, 85, 104, 0.3);
}

.network-name {
  font-size: 1.1rem;
  color: var(--reach-amber);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: var(--space-md);
}

.network-details {
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

.network-actions {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.network-actions .btn {
  font-size: 0.75rem;
  padding: var(--space-xs) var(--space-sm);
}
</style>
