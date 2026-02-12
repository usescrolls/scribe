<template>
  <div class="skill-card" :class="{ compact: mode === 'compact', selected: selectable && selected }" @click="selectable ? $emit('toggle-select', skill.name) : $emit('detail', skill)">
    <div class="skill-main">
      <input
        v-if="selectable"
        type="checkbox"
        class="select-checkbox"
        :checked="selected"
        @click.stop="$emit('toggle-select', skill.name)"
      />
      <span class="source-badge">{{ skill.sourceType }}</span>
      <span class="name">{{ skill.name }}</span>
      <span v-if="skill.description" class="description">{{ truncatedDescription }}</span>
    </div>
    <div class="skill-meta" v-if="skillWorkspaces && skillWorkspaces.length > 0">
      <span
        v-for="ws in skillWorkspaces"
        :key="ws"
        class="ws-badge"
      >{{ ws }}</span>
    </div>
    <div class="skill-right">
      <!-- Workspace picker dropdown -->
      <div v-if="showWorkspacePicker" class="ws-picker-wrapper">
        <button
          class="action-btn ws-picker-btn"
          @click.stop="togglePicker"
          title="Add to workspace"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          <span class="btn-label">Workspace</span>
        </button>
        <div v-if="pickerOpen" class="ws-picker-dropdown">
          <label
            v-for="ws in allWorkspaces"
            :key="ws.name"
            class="ws-picker-item"
            @click.stop
          >
            <input
              type="checkbox"
              :checked="skillWorkspaces?.includes(ws.name)"
              @change="handleWorkspaceToggle(ws.name, ($event.target as HTMLInputElement).checked)"
            />
            <span class="ws-picker-name">{{ ws.name }}</span>
          </label>
        </div>
      </div>
      <button
        v-if="showAdd"
        class="action-btn add-btn"
        @click.stop="$emit('add', skill.name)"
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
        @click.stop="$emit('remove', skill.name)"
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
        @click.stop="$emit('uninstall', skill.name)"
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { SkillInfo, WorkspaceInfo } from '../types/skill'

const props = withDefaults(defineProps<{
  skill: SkillInfo
  mode?: 'compact' | 'normal'
  showUninstall?: boolean
  showRemove?: boolean
  showAdd?: boolean
  showWorkspacePicker?: boolean
  selectable?: boolean
  selected?: boolean
  skillWorkspaces?: string[]
  allWorkspaces?: WorkspaceInfo[]
}>(), {
  mode: 'compact',
  showUninstall: false,
  showRemove: false,
  showAdd: false,
  showWorkspacePicker: false,
  selectable: false,
  selected: false,
  skillWorkspaces: () => [],
  allWorkspaces: () => []
})

const emit = defineEmits<{
  detail: [skill: SkillInfo]
  uninstall: [name: string]
  remove: [name: string]
  add: [name: string]
  'add-to-workspace': [skillName: string, workspaceName: string]
  'remove-from-workspace': [skillName: string, workspaceName: string]
  'toggle-select': [name: string]
}>()

const pickerOpen = ref(false)

function togglePicker() {
  pickerOpen.value = !pickerOpen.value
}

function handleWorkspaceToggle(workspaceName: string, checked: boolean) {
  if (checked) {
    emit('add-to-workspace', props.skill.name, workspaceName)
  } else {
    emit('remove-from-workspace', props.skill.name, workspaceName)
  }
}

function closePicker() {
  pickerOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', closePicker)
})

onUnmounted(() => {
  document.removeEventListener('click', closePicker)
})

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
  cursor: pointer;
}

.skill-card:hover {
  border-color: var(--accent-color);
}

.skill-card.selected {
  border-color: var(--accent-color);
  background-color: rgba(0, 113, 227, 0.06);
}

.select-checkbox {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
  accent-color: var(--accent-color);
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

.skill-meta {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.ws-badge {
  padding: 0.0625rem 0.3rem;
  background-color: rgba(0, 113, 227, 0.1);
  color: var(--accent-color);
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 500;
  white-space: nowrap;
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

.add-btn:hover,
.ws-picker-btn:hover {
  background-color: var(--accent-color);
  color: white;
}

/* Workspace picker dropdown */
.ws-picker-wrapper {
  position: relative;
}

.ws-picker-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: 10;
  min-width: 160px;
  margin-top: 0.25rem;
  padding: 0.25rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.ws-picker-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.5rem;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.ws-picker-item:hover {
  background-color: var(--bg-primary);
}

.ws-picker-item input[type="checkbox"] {
  width: 0.75rem;
  height: 0.75rem;
  cursor: pointer;
  flex-shrink: 0;
}

.ws-picker-name {
  font-size: 0.75rem;
  color: var(--text-primary);
  white-space: nowrap;
}
</style>
