<template>
  <div class="metric-chart hud-corners">
    <div class="chart-header">
      <h3 class="chart-title">{{ title }}</h3>
      <span class="chart-subtitle">{{ subtitle }}</span>
    </div>
    <div class="chart-container">
      <Line v-if="chartData" :data="chartData" :options="chartOptions" />
      <div v-else class="chart-empty">
        <span>No data available</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  subtitle: {
    type: String,
    default: ''
  },
  data: {
    type: Array,
    default: () => []
  },
  labels: {
    type: Array,
    default: () => []
  },
  color: {
    type: String,
    default: '#22d3ee' // cyan
  },
  unit: {
    type: String,
    default: '%'
  }
})

const chartData = computed(() => {
  if (!props.data || props.data.length === 0) return null

  return {
    labels: props.labels,
    datasets: [
      {
        label: props.title,
        data: props.data,
        borderColor: props.color,
        backgroundColor: `${props.color}20`,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 4,
        borderWidth: 2
      }
    ]
  }
})

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: '#2d3548',
      titleColor: '#f6a623',
      bodyColor: '#e2e8f0',
      borderColor: 'rgba(74, 85, 104, 0.5)',
      borderWidth: 1,
      padding: 12,
      displayColors: false,
      callbacks: {
        label: (context) => {
          return `${context.parsed.y.toFixed(1)}${props.unit}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: 'rgba(74, 85, 104, 0.2)',
        drawBorder: false
      },
      ticks: {
        color: '#94a3b8',
        font: {
          family: 'JetBrains Mono',
          size: 10
        }
      }
    },
    y: {
      beginAtZero: true,
      max: 100,
      grid: {
        color: 'rgba(74, 85, 104, 0.2)',
        drawBorder: false
      },
      ticks: {
        color: '#94a3b8',
        font: {
          family: 'JetBrains Mono',
          size: 10
        },
        callback: (value) => `${value}${props.unit}`
      }
    }
  },
  interaction: {
    intersect: false,
    mode: 'index'
  }
}))
</script>

<style scoped>
.metric-chart {
  padding: var(--space-lg);
  background-color: var(--reach-steel);
  border: 1px solid rgba(74, 85, 104, 0.3);
  border-radius: var(--radius-md);
}

.chart-header {
  margin-bottom: var(--space-md);
}

.chart-title {
  font-size: 1rem;
  color: var(--reach-amber);
  margin-bottom: var(--space-xs);
}

.chart-subtitle {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.chart-container {
  height: 250px;
  position: relative;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
}
</style>
