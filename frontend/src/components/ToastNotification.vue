<template>
  <Transition name="toast">
    <div v-if="visible" :class="['toast', `toast-${type}`]" @click="close">
      <span class="toast-message">{{ message }}</span>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(defineProps<{
  message: string
  type?: 'success' | 'error' | 'info'
  duration?: number
}>(), {
  type: 'info',
  duration: 4000,
})

const emit = defineEmits<{
  close: []
}>()

const visible = ref(false)
let timer: ReturnType<typeof setTimeout> | null = null

function close() {
  visible.value = false
  setTimeout(() => emit('close'), 200)
}

onMounted(() => {
  // Show on next tick for transition to work
  requestAnimationFrame(() => {
    visible.value = true
  })
  timer = setTimeout(close, props.duration)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})
</script>

<style scoped>
.toast {
  position: fixed;
  bottom: 1.5rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  max-width: 90%;
  text-align: center;
}

.toast-success {
  background-color: var(--success-color);
  color: white;
}

.toast-error {
  background-color: var(--danger-color);
  color: white;
}

.toast-info {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.2s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(0.5rem);
}
</style>
