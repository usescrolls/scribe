<template>
  <div class="skill-card" :class="{ compact: mode === 'compact' }">
    <div class="skill-main">
      <span class="source-badge">{{ skill.sourceType }}</span>
      <span class="name">{{ skill.name }}</span>
      <span v-if="skill.description" class="description">{{ truncatedDescription }}</span>
    </div>
    <div class="skill-right">
      <button
        v-if="showAdd"
        class="action-btn add-btn"
        @click="$emit('add', skill.name)"
        title="Add to workspace"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span class="btn-label">Add</span>
      </button>
      <button
        v-if="showRemove"
        class="action-btn remove-btn"
        @click="$emit('remove', skill.name)"
        title="Remove from workspace"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        <span class="btn-label">Remove</span>
      </button>
      <button
        v-if="showUninstall"
        class="action-btn uninstall-btn"
        @click="$emit('uninstall', skill.name)"
        title="Uninstall skill"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3 6 5 6 21 6"></polyline>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
        </svg>
        <span class="btn-label">Uninstall</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SkillInfo } from '../types/skill'

const props = withDefaults(defineProps<{
  skill: SkillInfo
  mode?: 'compact' | 'normal'
  showUninstall?: boolean
  showRemove?: boolean
  showAdd?: boolean
}>(), {
  mode: 'compact',
  showUninstall: false,
  showRemove: false,
  showAdd: false
})

defineEmits<{
  uninstall: [name: string]
  remove: [name: string]
  add: [name: string]
}>()

const truncatedDescription = computed(() => {
  if (!props.skill.description) return ''
  const maxLen = 60
  if (props.skill.description.length <= maxLen) return props.skill.description
  return props.skill.description.slice(0, maxLen).trim() + '...'
})

</script>

<style scoped>
.skill-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  transition: border-color 0.15s;
}

.skill-card:hover {
  border-color: var(--accent-color);
}

.skill-main {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.source-badge {
  flex-shrink: 0;
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 3px;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.name {
  flex-shrink: 0;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.description {
  flex: 1;
  min-width: 0;
  font-size: 0.75rem;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-right {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.action-btn {
  height: 1.5rem;
  padding: 0 0.5rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  transition: all 0.15s;
  opacity: 0;
}

.btn-label {
  font-size: 0.6875rem;
  font-weight: 500;
}

.skill-card:hover .action-btn {
  opacity: 1;
}

.uninstall-btn:hover,
.remove-btn:hover {
  background-color: var(--danger-color);
  color: white;
}

.add-btn:hover {
  background-color: var(--accent-color);
  color: white;
}
</style>
