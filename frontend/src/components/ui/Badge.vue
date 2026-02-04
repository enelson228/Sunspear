<template>
  <span :class="['badge', `badge-${variant}`]">
    <div v-if="showDot" :class="['status-dot', variant, { 'status-pulse': pulse }]"></div>
    <slot />
  </span>
</template>

<script setup>
defineProps({
  variant: {
    type: String,
    default: 'online',
    validator: (value) => ['online', 'offline', 'warning'].includes(value)
  },
  showDot: {
    type: Boolean,
    default: true
  },
  pulse: {
    type: Boolean,
    default: false
  }
})
</script>
