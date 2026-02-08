<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="modelValue"
        class="modal-overlay"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        @click.self="close"
        @keydown.escape="close"
      >
        <div ref="modalRef" class="modal hud-corners" :class="{ 'modal-wide': wide }" tabindex="-1">
          <div :id="titleId" class="modal-header">
            <slot name="header">{{ title }}</slot>
          </div>
          <div class="modal-body">
            <slot />
          </div>
          <div v-if="$slots.footer" class="modal-footer flex justify-end gap-md mt-lg">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, useId } from 'vue'

const props = defineProps({
  modelValue: Boolean,
  title: String,
  wide: Boolean
})

const emit = defineEmits(['update:modelValue'])
const modalRef = ref(null)
const titleId = `modal-title-${useId()}`

function close() {
  emit('update:modelValue', false)
}

// Focus the modal when it opens
watch(() => props.modelValue, (open) => {
  if (open) {
    requestAnimationFrame(() => {
      modalRef.value?.focus()
    })
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-base);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-footer {
  margin-top: var(--space-lg);
  padding-top: var(--space-lg);
  border-top: 1px solid rgba(74, 85, 104, 0.3);
}
</style>
