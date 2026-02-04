<template>
  <div :class="['card', { 'hud-corners': showCorners }]">
    <slot />
  </div>
</template>

<script setup>
defineProps({
  showCorners: {
    type: Boolean,
    default: true
  }
})
</script>
