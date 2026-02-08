<template>
  <div class="input-wrapper">
    <label v-if="label" :for="inputId" class="label">{{ label }}</label>
    <input
      :id="inputId"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :aria-invalid="error ? 'true' : undefined"
      :aria-describedby="error ? errorId : undefined"
      class="input"
      :class="{ 'input-error': error }"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <p v-if="error" :id="errorId" class="input-error-text" role="alert">{{ error }}</p>
  </div>
</template>

<script setup>
import { useId } from 'vue'

const inputId = `input-${useId()}`
const errorId = `input-error-${useId()}`

defineProps({
  modelValue: [String, Number],
  label: String,
  type: {
    type: String,
    default: 'text'
  },
  placeholder: String,
  disabled: Boolean,
  error: String
})

defineEmits(['update:modelValue'])
</script>

<style scoped>
.input-wrapper {
  width: 100%;
}
.input-error {
  border-color: var(--status-error);
}
.input-error-text {
  color: var(--status-error);
  font-size: 0.75rem;
  margin-top: var(--space-xs);
}
</style>
