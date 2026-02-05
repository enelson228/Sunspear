<template>
  <div class="metric-gauge hud-corners">
    <div class="gauge-header">
      <div class="gauge-title">
        <span :class="['status-dot', statusClass, 'status-pulse']"></span>
        <span class="font-mono">{{ title }}</span>
      </div>
      <span class="gauge-value" :class="statusClass">{{ displayValue }}</span>
    </div>
    <div class="gauge-bar">
      <div
        class="gauge-fill"
        :class="statusClass"
        :style="{ width: `${percentage}%` }"
      ></div>
    </div>
    <div class="gauge-footer">
      <span class="gauge-label">{{ label }}</span>
      <span class="gauge-percent font-mono">{{ percentage.toFixed(1) }}%</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  value: {
    type: Number,
    required: true
  },
  max: {
    type: Number,
    default: 100
  },
  unit: {
    type: String,
    default: '%'
  },
  label: {
    type: String,
    default: ''
  },
  warningThreshold: {
    type: Number,
    default: 75
  },
  criticalThreshold: {
    type: Number,
    default: 90
  }
})

const percentage = computed(() => {
  return Math.min((props.value / props.max) * 100, 100)
})

const displayValue = computed(() => {
  if (props.unit === '%') {
    return `${props.value.toFixed(1)}%`
  }
  return `${props.value} ${props.unit}`
})

const statusClass = computed(() => {
  if (percentage.value >= props.criticalThreshold) return 'critical'
  if (percentage.value >= props.warningThreshold) return 'warning'
  return 'online'
})
</script>

<style scoped>
.metric-gauge {
  padding: var(--space-lg);
  background-color: var(--reach-steel);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-md);
}

.gauge-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
}

.gauge-title {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.gauge-value {
  font-family: var(--font-display);
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.gauge-value.online {
  color: var(--reach-cyan);
}

.gauge-value.warning {
  color: var(--reach-amber);
}

.gauge-value.critical {
  color: var(--reach-orange);
}

.gauge-bar {
  height: 8px;
  background-color: var(--reach-slate);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: var(--space-sm);
}

.gauge-fill {
  height: 100%;
  transition: width var(--transition-slow), background-color var(--transition-base);
  border-radius: 4px;
}

.gauge-fill.online {
  background: linear-gradient(90deg, var(--reach-cyan), rgba(34, 211, 238, 0.6));
  box-shadow: 0 0 10px var(--reach-cyan);
}

.gauge-fill.warning {
  background: linear-gradient(90deg, var(--reach-amber), rgba(246, 166, 35, 0.6));
  box-shadow: 0 0 10px var(--reach-amber);
}

.gauge-fill.critical {
  background: linear-gradient(90deg, var(--reach-orange), rgba(232, 93, 4, 0.6));
  box-shadow: 0 0 10px var(--reach-orange);
}

.gauge-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.gauge-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.gauge-percent {
  font-size: 0.875rem;
  color: var(--text-secondary);
}
</style>
