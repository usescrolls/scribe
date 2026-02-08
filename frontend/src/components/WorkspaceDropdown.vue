<template>
  <div class="workspace-dropdown" ref="dropdownRef">
    <button class="dropdown-trigger" @click="toggleOpen">
      <span class="trigger-label">Workspace:</span>
      <span class="active-name">{{ activeWorkspace || 'default' }}</span>
      <svg class="chevron" :class="{ open: isOpen }" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polyline points="6 9 12 15 18 9"></polyline>
      </svg>
    </button>

    <div v-if="isOpen" class="dropdown-panel">
      <div class="workspace-list">
        <div
          v-for="ws in workspaces"
          :key="ws.name"
          :class="['workspace-item', { active: ws.isActive }]"
          @click="handleSwitch(ws.name)"
        >
          <span class="workspace-icon">
            <span v-if="ws.isActive" class="checkmark">&#10003;</span>
            <span v-else class="empty">&#9675;</span>
          </span>
          <span class="workspace-name">{{ ws.name }}</span>
          <span class="skill-count">{{ ws.skills.length }}</span>
          <button
            v-if="ws.name !== 'default'"
            class="delete-btn"
            @click.stop="handleDelete(ws.name)"
            title="Delete workspace"
          >×</button>
        </div>
      </div>

      <div class="dropdown-divider"></div>

      <div v-if="isAdding" class="add-form">
        <input
          ref="addInput"
          v-model="newName"
          type="text"
          placeholder="Workspace name"
          @keyup.enter="handleCreate"
          @keyup.escape="cancelAdd"
        />
        <button class="create-btn" @click="handleCreate" :disabled="!newName.trim()">Create</button>
      </div>
      <button v-else class="add-workspace-btn" @click="startAdd">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"></line>
          <line x1="5" y1="12" x2="19" y2="12"></line>
        </svg>
        New workspace
      </button>

      <!-- Delete confirmation -->
      <div v-if="pendingDelete" class="delete-confirm">
        <p>Delete "<strong>{{ pendingDelete }}</strong>"?</p>
        <p class="confirm-note">Skills will remain installed.</p>
        <div class="confirm-actions">
          <button class="cancel-btn" @click="pendingDelete = null">Cancel</button>
          <button class="confirm-delete-btn" @click="confirmDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from 'vue'
import { useWorkspaces } from '../composables/useWorkspaces'

const { workspaces, activeWorkspace, switchWorkspace, createWorkspace, deleteWorkspace } = useWorkspaces()

const dropdownRef = ref<HTMLElement | null>(null)
const addInput = ref<HTMLInputElement | null>(null)
const isOpen = ref(false)
const isAdding = ref(false)
const newName = ref('')
const pendingDelete = ref<string | null>(null)

function toggleOpen() {
  isOpen.value = !isOpen.value
  if (!isOpen.value) {
    resetState()
  }
}

function resetState() {
  isAdding.value = false
  newName.value = ''
  pendingDelete.value = null
}

async function handleSwitch(name: string) {
  if (name === activeWorkspace.value) return
  await switchWorkspace(name)
  isOpen.value = false
  resetState()
}

function handleDelete(name: string) {
  pendingDelete.value = name
}

async function confirmDelete() {
  if (!pendingDelete.value) return
  const name = pendingDelete.value
  pendingDelete.value = null
  await deleteWorkspace(name)
}

function startAdd() {
  isAdding.value = true
  newName.value = ''
  nextTick(() => addInput.value?.focus())
}

function cancelAdd() {
  isAdding.value = false
  newName.value = ''
}

async function handleCreate() {
  const name = newName.value.trim()
  if (!name) return
  const success = await createWorkspace(name)
  if (success) {
    cancelAdd()
  }
}

function handleClickOutside(e: MouseEvent) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false
    resetState()
  }
}

onMounted(() => {
  document.addEventListener('mousedown', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('mousedown', handleClickOutside)
})
</script>

<style scoped>
.workspace-dropdown {
  position: relative;
  -webkit-app-region: no-drag;
}

.dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.dropdown-trigger:hover {
  border-color: var(--accent-color);
}

.trigger-label {
  color: var(--text-secondary);
  font-weight: 400;
}

.active-name {
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chevron {
  flex-shrink: 0;
  transition: transform 0.15s;
}

.chevron.open {
  transform: rotate(180deg);
}

.dropdown-panel {
  position: absolute;
  top: calc(100% + 0.375rem);
  right: 0;
  z-index: 50;
  min-width: 220px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  overflow: hidden;
}

.workspace-list {
  max-height: 240px;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0.25rem 0;
}

.workspace-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4375rem 0.75rem;
  cursor: pointer;
  transition: background-color 0.15s;
  font-size: 0.8125rem;
}

.workspace-item:hover {
  background-color: var(--bg-primary);
}

.workspace-item.active {
  background-color: rgba(0, 113, 227, 0.08);
}

.workspace-icon {
  width: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6875rem;
  flex-shrink: 0;
}

.checkmark {
  color: #34c759;
}

.empty {
  color: var(--text-secondary);
}

.workspace-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

.skill-count {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.delete-btn {
  width: 1.125rem;
  height: 1.125rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.875rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 3px;
  opacity: 0;
  transition: all 0.15s;
  flex-shrink: 0;
}

.workspace-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background-color: var(--danger-color);
  color: white;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--border-color);
}

.add-workspace-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.15s;
}

.add-workspace-btn:hover {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.add-form {
  display: flex;
  gap: 0.375rem;
  padding: 0.5rem 0.75rem;
}

.add-form input {
  flex: 1;
  padding: 0.3125rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.75rem;
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.add-form input:focus {
  outline: none;
  border-color: var(--accent-color);
}

.create-btn {
  padding: 0.3125rem 0.625rem;
  border: none;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
}

.create-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.delete-confirm {
  padding: 0.625rem 0.75rem;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.delete-confirm p {
  margin: 0 0 0.375rem 0;
  font-size: 0.75rem;
  color: var(--text-primary);
}

.confirm-note {
  color: var(--text-secondary) !important;
  font-size: 0.6875rem !important;
}

.confirm-actions {
  display: flex;
  gap: 0.375rem;
  justify-content: flex-end;
}

.cancel-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.6875rem;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
}

.confirm-delete-btn {
  padding: 0.25rem 0.5rem;
  border: none;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 500;
  background-color: var(--danger-color);
  color: white;
  cursor: pointer;
}
</style>
