<template>
  <img
    v-if="avatarUrl && !failed"
    :src="avatarUrl"
    class="source-avatar"
    alt=""
    @error="failed = true"
  />
  <div v-else-if="sourceType === 'github'" class="source-avatar-placeholder">
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
      <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  source: string
  sourceType: string
  isPrivate?: boolean
}>()

const failed = ref(false)

// Reset failed state when the source changes
watch(() => props.source, () => {
  failed.value = false
})

const avatarUrl = computed<string | null>(() => {
  if (props.sourceType !== 'github') return null
  if (props.isPrivate) return null
  const parts = props.source.split('/')
  if (parts.length < 2) return null
  return `https://github.com/${parts[0]}.png?size=48`
})
</script>

<style scoped>
.source-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  flex-shrink: 0;
  object-fit: cover;
}

.source-avatar-placeholder {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}
</style>
