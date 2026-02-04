<template>
  <div class="status-panel hud-corners">
    <div class="status-panel-header">
      <div :class="['status-dot', statusClass, { 'status-pulse': pulse }]"></div>
      <span>{{ title }}</span>
    </div>
    <div class="status-panel-value">{{ value }}</div>
    <div class="status-panel-label">{{ label }}</div>
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  value: {
    type: [String, Number],
    required: true
  },
  label: {
    type: String,
    default: ''
  },
  status: {
    type: String,
    default: 'online',
    validator: (value) => ['online', 'offline', 'warning'].includes(value)
  },
  pulse: {
    type: Boolean,
    default: true
  }
})

const statusClass = computed(() => props.status)
</script>

<script>
import { computed } from 'vue'
export default {
  name: 'StatusPanel'
}
</script>
