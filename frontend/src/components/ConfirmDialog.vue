<template>
  <Teleport to="body">
    <Transition name="modal" appear @after-leave="onAfterLeave">
      <div v-if="visible" class="confirm-backdrop" @click.self="handleCancel">
        <div class="confirm-modal">
          <div class="confirm-header">
            <h3>{{ title }}</h3>
          </div>
          <div class="confirm-body">
            <p class="confirm-message">{{ message }}</p>
          </div>
          <div class="confirm-footer">
            <button class="btn-cancel" @click="handleCancel">Cancel</button>
            <button
              :class="['btn-confirm', { danger: danger }]"
              @click="handleConfirm"
            >
              {{ confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

defineProps<{
  title: string
  message: string
  confirmLabel: string
  danger?: boolean
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const visible = ref(true)
let action: 'confirm' | 'cancel' = 'cancel'

function handleConfirm() {
  action = 'confirm'
  visible.value = false
}

function handleCancel() {
  action = 'cancel'
  visible.value = false
}

function onAfterLeave() {
  if (action === 'confirm') {
    emit('confirm')
  } else {
    emit('cancel')
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    handleCancel()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.confirm-backdrop {
  position: fixed;
  inset: 0;
  z-index: 200;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.confirm-modal {
  width: 340px;
  max-width: calc(100vw - 2rem);
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.confirm-header {
  padding: 1rem 1.25rem 0;
}

.confirm-header h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
}

.confirm-body {
  padding: 0.5rem 1.25rem 1rem;
}

.confirm-message {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.confirm-footer {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.btn-cancel {
  padding: 0.4375rem 0.875rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.btn-cancel:hover {
  background-color: var(--bg-secondary);
}

.btn-confirm {
  padding: 0.4375rem 0.875rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-confirm:hover {
  opacity: 0.9;
}

.btn-confirm.danger {
  background-color: var(--danger-color);
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .confirm-modal,
.modal-leave-active .confirm-modal {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .confirm-modal {
  transform: scale(0.95);
  opacity: 0;
}

.modal-leave-to .confirm-modal {
  transform: scale(0.95);
  opacity: 0;
}
</style>
