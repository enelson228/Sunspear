<template>
  <button
    :class="['btn', variant && `btn-${variant}`, { 'accent-bar': showAccent }]"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="spinner" style="width: 16px; height: 16px; border-width: 2px; margin-right: 8px;"></span>
    <slot />
  </button>
</template>

<script setup>
defineProps({
  variant: {
    type: String,
    default: 'primary',
    validator: (value) => ['primary', 'secondary', 'danger'].includes(value)
  },
  disabled: Boolean,
  loading: Boolean,
  showAccent: {
    type: Boolean,
    default: false
  }
})

defineEmits(['click'])
</script>

<style scoped>
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
