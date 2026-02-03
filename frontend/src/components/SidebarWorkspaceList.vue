<template>
  <div class="workspace-list-panel">
    <div class="panel-header">
      <h3>Workspaces</h3>
      <span class="summary">{{ workspaces.length }} total</span>
      <button
        class="add-btn"
        @click="toggleAddMode"
        :title="isAddMode ? 'Cancel' : 'Add workspace'"
      >
        {{ isAddMode ? '×' : '+' }}
      </button>
    </div>
    <div v-if="isAddMode" class="add-workspace-form">
      <input
        ref="newNameInput"
        v-model="newWorkspaceName"
        type="text"
        placeholder="Workspace name"
        @keyup.enter="handleCreate"
        @keyup.escape="cancelAdd"
      />
      <button class="create-btn" @click="handleCreate" :disabled="!newWorkspaceName.trim()">
        Create
      </button>
    </div>

    <!-- Switch confirmation dialog -->
    <div v-if="pendingSwitchTo" class="confirm-dialog">
      <p class="confirm-message">Switch to "<strong>{{ pendingSwitchTo }}</strong>"?</p>
      <label class="dont-ask-label">
        <input type="checkbox" v-model="dontAskAgain" />
        <span>Don't ask me again</span>
      </label>
      <div class="confirm-actions">
        <button class="cancel-btn" @click="cancelSwitch">Cancel</button>
        <button class="confirm-btn" @click="confirmSwitch">Switch</button>
      </div>
    </div>

    <!-- Delete confirmation dialog -->
    <div v-if="pendingDeleteName" class="confirm-dialog delete-confirm">
      <p class="confirm-message">Delete "<strong>{{ pendingDeleteName }}</strong>"?</p>
      <p class="confirm-note">Skills will remain installed.</p>
      <div class="confirm-actions">
        <button class="cancel-btn" @click="cancelDelete">Cancel</button>
        <button class="delete-confirm-btn" @click="confirmDelete">Delete</button>
      </div>
    </div>

    <div class="workspace-list">
      <div
        v-for="ws in workspaces"
        :key="ws.name"
        :class="['workspace-item', { active: ws.isActive }]"
        @click="handleWorkspaceClick(ws.name)"
      >
        <div class="workspace-icon">
          <span v-if="ws.isActive" class="checkmark">&#10003;</span>
          <span v-else class="empty">&#9675;</span>
        </div>
        <div class="workspace-info">
          <span class="workspace-name">{{ ws.name }}</span>
          <span class="skill-count">
            {{ ws.skills.length }} skill{{ ws.skills.length !== 1 ? 's' : '' }}
          </span>
        </div>
        <button
          v-if="ws.name !== 'default'"
          class="delete-btn"
          @click.stop="handleDeleteClick(ws.name)"
          title="Delete workspace"
        >
          ×
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useWorkspaces } from '../composables/useWorkspaces'

const SKIP_CONFIRM_KEY = 'scribe-skip-workspace-switch-confirm'

const { workspaces, activeWorkspace, loading, switchWorkspace, createWorkspace, deleteWorkspace } = useWorkspaces()

const isAddMode = ref(false)
const newWorkspaceName = ref('')
const newNameInput = ref<HTMLInputElement | null>(null)

// Switch confirmation state
const pendingSwitchTo = ref<string | null>(null)
const dontAskAgain = ref(false)

// Delete confirmation state
const pendingDeleteName = ref<string | null>(null)

function shouldSkipConfirm(): boolean {
  return localStorage.getItem(SKIP_CONFIRM_KEY) === 'true'
}

async function handleWorkspaceClick(workspaceName: string) {
  if (workspaceName === activeWorkspace.value || loading.value) {
    return
  }

  if (shouldSkipConfirm()) {
    await switchWorkspace(workspaceName)
  } else {
    pendingSwitchTo.value = workspaceName
    dontAskAgain.value = false
  }
}

function cancelSwitch() {
  pendingSwitchTo.value = null
  dontAskAgain.value = false
}

async function confirmSwitch() {
  if (!pendingSwitchTo.value) return

  if (dontAskAgain.value) {
    localStorage.setItem(SKIP_CONFIRM_KEY, 'true')
  }

  const target = pendingSwitchTo.value
  pendingSwitchTo.value = null
  dontAskAgain.value = false

  await switchWorkspace(target)
}

function toggleAddMode() {
  isAddMode.value = !isAddMode.value
  if (isAddMode.value) {
    newWorkspaceName.value = ''
    nextTick(() => {
      newNameInput.value?.focus()
    })
  }
}

function cancelAdd() {
  isAddMode.value = false
  newWorkspaceName.value = ''
}

async function handleCreate() {
  const name = newWorkspaceName.value.trim()
  if (!name) return

  const success = await createWorkspace(name)
  if (success) {
    cancelAdd()
  }
}

function handleDeleteClick(workspaceName: string) {
  pendingDeleteName.value = workspaceName
}

function cancelDelete() {
  pendingDeleteName.value = null
}

async function confirmDelete() {
  if (!pendingDeleteName.value) return

  const name = pendingDeleteName.value
  pendingDeleteName.value = null

  await deleteWorkspace(name)
}
</script>

<style scoped>
.workspace-list-panel {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
}

.panel-header h3 {
  font-size: 0.875rem;
  font-weight: 600;
  margin: 0;
}

.summary {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.add-btn {
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

.add-btn:hover {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.add-workspace-form {
  display: flex;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--border-color);
}

.add-workspace-form input {
  flex: 1;
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.8125rem;
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.add-workspace-form input:focus {
  outline: none;
  border-color: var(--accent-color);
}

.create-btn {
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: opacity 0.15s;
}

.create-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.create-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Confirmation dialogs */
.confirm-dialog {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--bg-primary);
}

.confirm-message {
  margin: 0 0 0.5rem 0;
  font-size: 0.8125rem;
  color: var(--text-primary);
}

.confirm-note {
  margin: 0 0 0.5rem 0;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.dont-ask-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
  cursor: pointer;
  margin-bottom: 0.75rem;
}

.dont-ask-label input[type="checkbox"] {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
}

.confirm-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.cancel-btn {
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.cancel-btn:hover {
  background-color: var(--bg-primary);
}

.confirm-btn {
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: opacity 0.15s;
}

.confirm-btn:hover {
  opacity: 0.9;
}

.delete-confirm-btn {
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--danger-color);
  color: white;
  cursor: pointer;
  transition: opacity 0.15s;
}

.delete-confirm-btn:hover {
  opacity: 0.9;
}

.workspace-list {
  max-height: 200px;
  overflow-y: auto;
}

.workspace-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: background-color 0.15s;
}

.workspace-item:hover {
  background-color: var(--bg-primary);
}

.workspace-item.active {
  background-color: rgba(0, 113, 227, 0.1);
}

.workspace-icon {
  width: 1.25rem;
  height: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  flex-shrink: 0;
}

.checkmark {
  color: #34c759;
}

.empty {
  color: var(--text-secondary);
}

.workspace-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
  flex: 1;
}

.workspace-name {
  font-size: 0.8125rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-count {
  font-size: 0.6875rem;
  color: var(--text-secondary);
}

.delete-btn {
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 4px;
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
</style>
