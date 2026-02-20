<template>
  <Teleport to="body">
    <Transition name="modal" appear @after-leave="onAfterLeave">
      <div v-if="visible" class="info-backdrop" @click.self="handleDismiss">
        <div class="info-modal">
          <div class="info-header">
            <h3>Workspace activated</h3>
            <button class="close-btn" @click="handleDismiss" aria-label="Close">&times;</button>
          </div>
          <div class="info-body">
            <p class="info-message">
              You switched to <strong>{{ switchInfoWorkspaceName }}</strong>. The skills in this workspace will only be available when you <strong>open a new coding agent chat window</strong>.
            </p>
            <p class="info-callout">
              On Cursor, you'll need to create a new chat tab for the updated skills to load.
            </p>
          </div>
          <div class="info-footer">
            <label class="dont-show-label">
              <input type="checkbox" v-model="dontShowAgain" />
              <span>Don't show this again</span>
            </label>
            <button class="btn-got-it" @click="handleDismiss">Got it</button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { showSwitchInfoModal, switchInfoWorkspaceName, dismissSwitchInfo } from '../composables/useWorkspaces'

const visible = ref(false)
const dontShowAgain = ref(false)
let pendingDismiss = false

watch(showSwitchInfoModal, (val) => {
  if (val) {
    visible.value = true
    dontShowAgain.value = false
    pendingDismiss = false
  }
})

function handleDismiss() {
  pendingDismiss = true
  visible.value = false
}

function onAfterLeave() {
  if (pendingDismiss) {
    dismissSwitchInfo(dontShowAgain.value)
    pendingDismiss = false
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && visible.value) {
    handleDismiss()
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
.info-backdrop {
  position: fixed;
  inset: 0;
  z-index: 200;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.info-modal {
  width: 400px;
  max-width: calc(100vw - 2rem);
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  overflow: hidden;
}

.info-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem 0;
}

.info-header h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
}

.close-btn {
  width: 1.5rem;
  height: 1.5rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.15s;
}

.close-btn:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.info-body {
  padding: 0.75rem 1.25rem 1rem;
}

.info-message {
  margin: 0 0 0.75rem 0;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.6;
}

.info-callout {
  margin: 0;
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.5;
  padding: 0.5rem 0.75rem;
  background-color: var(--bg-secondary);
  border-radius: 6px;
  border-left: 3px solid var(--accent-color);
}

.info-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.dont-show-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
  cursor: pointer;
}

.dont-show-label input[type="checkbox"] {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
}

.btn-got-it {
  padding: 0.4375rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-got-it:hover {
  opacity: 0.9;
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .info-modal,
.modal-leave-active .info-modal {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .info-modal {
  transform: scale(0.95);
  opacity: 0;
}

.modal-leave-to .info-modal {
  transform: scale(0.95);
  opacity: 0;
}
</style>
