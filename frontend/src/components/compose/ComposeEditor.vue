<template>
  <div class="compose-editor">
    <label v-if="label" class="label">{{ label }}</label>
    <textarea
      :value="modelValue"
      @input="$emit('update:modelValue', $event.target.value)"
      class="yaml-input"
      :placeholder="placeholder"
      rows="16"
    ></textarea>

    <div class="validation-section">
      <Button
        variant="secondary"
        size="sm"
        @click="$emit('validate')"
        :loading="validating"
      >
        {{ validating ? 'VALIDATING...' : 'VALIDATE YAML' }}
      </Button>

      <div v-if="validationResult" class="validation-result" :class="`validation-${validationResult.type}`">
        {{ validationResult.message }}
      </div>
    </div>
  </div>
</template>

<script setup>
import Button from '@/components/ui/Button.vue'

defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  label: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: 'version: "3.8"\nservices:\n  app:\n    image: nginx:latest\n    ports:\n      - "80:80"'
  },
  validating: {
    type: Boolean,
    default: false
  },
  validationResult: {
    type: Object,
    default: null
  }
})

defineEmits(['update:modelValue', 'validate'])
</script>

<style scoped>
.compose-editor {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.label {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-secondary);
}

.yaml-input {
  width: 100%;
  min-height: 400px;
  padding: var(--space-md);
  background-color: var(--reach-slate);
  border: 1px solid rgba(74, 85, 104, 0.5);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: 0.875rem;
  resize: vertical;
}

.yaml-input:focus {
  outline: none;
  border-color: var(--reach-amber);
  box-shadow: 0 0 0 2px rgba(246, 166, 35, 0.2);
}

.yaml-input::placeholder {
  color: var(--text-muted);
}

.validation-section {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.validation-result {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  padding: var(--space-sm) var(--space-md);
  border-radius: var(--radius-sm);
}

.validation-success {
  color: var(--reach-cyan);
  border: 1px solid rgba(34, 211, 238, 0.3);
  background-color: rgba(34, 211, 238, 0.1);
}

.validation-error {
  color: var(--reach-orange);
  border: 1px solid rgba(255, 119, 0, 0.3);
  background-color: rgba(255, 119, 0, 0.1);
}
</style>
