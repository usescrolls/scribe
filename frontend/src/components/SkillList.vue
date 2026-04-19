<template>
  <div class="skill-list">
    <ConfirmDialog
      v-if="confirmRemoveName"
      title="Remove from Workspace"
      :message="`Remove &quot;${confirmRemoveName}&quot; from this workspace?`"
      confirm-label="Remove"
      :danger="true"
      @confirm="executeRemove"
      @cancel="confirmRemoveName = null"
    />
    <ConfirmDialog
      v-if="confirmBulkRemove"
      title="Remove from Workspace"
      :message="`Remove ${selectedSkills.size} skill${selectedSkills.size !== 1 ? 's' : ''} from this workspace?`"
      confirm-label="Remove All"
      :danger="true"
      @confirm="executeBulkRemove"
      @cancel="confirmBulkRemove = false"
    />
    <ConfirmDialog
      v-if="confirmRemoveGroup"
      title="Remove from Workspace"
      :message="`Remove all ${confirmRemoveGroup.skills.filter(s => !s.isSystem).length} skills from &quot;${confirmRemoveGroup.source}&quot; from this workspace?`"
      confirm-label="Remove All"
      :danger="true"
      @confirm="executeRemoveGroup"
      @cancel="confirmRemoveGroup = null"
    />
    <ToastNotification
      v-if="toast"
      :message="toast.message"
      :type="toast.type"
      @close="toast = null"
    />
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>Loading skills...</span>
    </div>
    <div v-else-if="error" class="error">
      <span>{{ error }}</span>
      <button class="btn-secondary" @click="fetchAll">Retry</button>
    </div>
    <EmptyState v-else-if="filteredSkills.length === 0" />
    <div v-else class="skills">
      <div class="skills-header">
        <span class="count">{{ filteredSkills.length }} skill{{ filteredSkills.length !== 1 ? 's' : '' }} in workspace</span>
        <div class="header-actions">
          <template v-if="selectionMode">
            <button class="btn-secondary btn-sm" @click="toggleSelectAll">
              {{ allSelected ? 'Deselect All' : 'Select All' }}
            </button>
            <button class="btn-secondary btn-sm" @click="exitSelectionMode">Cancel</button>
          </template>
          <button v-else-if="selectableSkills.length > 0" class="btn-secondary btn-sm" @click="enterSelectionMode">Select</button>
        </div>
      </div>

      <div v-for="group in groupedSkills" :key="group.key" class="source-group">
        <div class="group-header">
          <input
            v-if="selectionMode && hasSelectableSkills(group)"
            type="checkbox"
            class="group-checkbox"
            :checked="isGroupAllSelected(group)"
            :indeterminate="isGroupPartiallySelected(group)"
            @click.stop="toggleGroupSelection(group)"
          />
          <SourceAvatar :source="group.source" :source-type="group.sourceType" />
          <span class="group-badge">{{ group.sourceType }}</span>
          <a
            v-if="group.sourceUrl && isHttpUrl(group.sourceUrl)"
            class="group-source group-source-link"
            @click.prevent="Browser.OpenURL(group.sourceUrl!)"
          >
            {{ group.source }}
            <svg class="external-icon" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
              <polyline points="15 3 21 3 21 9"></polyline>
              <line x1="10" y1="14" x2="21" y2="3"></line>
            </svg>
          </a>
          <span v-else class="group-source">{{ group.source }}</span>
          <span class="group-count">{{ group.skills.length }}</span>
          <button
            v-if="!selectionMode && group.skills.filter(s => !s.isSystem).length > 1"
            class="group-action-btn group-remove-btn"
            @click="handleRemoveGroup(group)"
          >
            Remove all
          </button>
        </div>
        <div class="skills-list">
          <SkillCard
            v-for="skill in group.skills"
            :key="skill.name"
            :skill="skill"
            :show-remove="!selectionMode && !skill.isSystem"
            :selectable="selectionMode && !skill.isSystem"
            :selected="selectedSkills.has(skill.name)"
            @detail="handleDetail"
            @remove="handleRemove"
            @toggle-select="toggleSkillSelection"
          />
        </div>
      </div>

      <!-- Bulk action bar -->
      <div v-if="selectionMode && selectedSkills.size > 0" class="bulk-action-bar">
        <span class="bulk-count">{{ selectedSkills.size }} skill{{ selectedSkills.size !== 1 ? 's' : '' }} selected</span>
        <button
          class="btn-danger btn-sm"
          :disabled="bulkRemoving"
          @click="confirmBulkRemove = true"
        >
          {{ bulkRemoving ? 'Removing...' : 'Remove from workspace' }}
        </button>
        <button class="btn-secondary btn-sm" @click="exitSelectionMode">Cancel</button>
      </div>
    </div>
    <SkillDetailModal
      v-if="detailSkill"
      :skill="detailSkill"
      @close="detailSkill = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { Browser, Events } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import SkillCard from './SkillCard.vue'
import SourceAvatar from './SourceAvatar.vue'
import EmptyState from './EmptyState.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import ToastNotification from './ToastNotification.vue'
import SkillDetailModal from './SkillDetailModal.vue'
import { sourceGroupKey } from '../utils/source'
import type { SkillInfo, WorkspaceInfo } from '../types/skill'

const skills = ref<SkillInfo[]>([])
const workspaces = ref<WorkspaceInfo[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const toast = ref<{ message: string; type: 'success' | 'error' | 'info' } | null>(null)
const selectedSkills = reactive(new Set<string>())
const selectionMode = ref(false)
const bulkRemoving = ref(false)
const confirmBulkRemove = ref(false)
let unsubscribeSkills: { (): void } | null = null
let unsubscribeWorkspace: { (): void } | null = null

const activeWorkspace = computed(() => {
  return workspaces.value.find(ws => ws.isActive)
})

const workspaceSkillNames = computed(() => {
  return new Set(activeWorkspace.value?.skills || [])
})

const filteredSkills = computed(() => {
  return skills.value.filter(skill => workspaceSkillNames.value.has(skill.name))
})

interface SourceGroup {
  key: string
  source: string
  sourceType: string
  sourceUrl?: string
  skills: SkillInfo[]
}

const groupedSkills = computed<SourceGroup[]>(() => {
  const groups = new Map<string, SourceGroup>()
  for (const skill of filteredSkills.value) {
    const key = sourceGroupKey(skill.source, skill.sourceType)
    if (!groups.has(key)) {
      groups.set(key, {
        key,
        source: skill.source || 'Unknown source',
        sourceType: skill.sourceType || 'local',
        sourceUrl: skill.sourceUrl,
        skills: [],
      })
    }
    groups.get(key)!.skills.push(skill)
  }
  return [...groups.values()]
})

function getScrollContainer(): HTMLElement | null {
  return document.querySelector('.main')
}

async function fetchAll() {
  const isInitialLoad = skills.value.length === 0
  const scrollContainer = getScrollContainer()
  const scrollTop = scrollContainer?.scrollTop ?? 0
  try {
    if (isInitialLoad) loading.value = true
    error.value = null
    const [skillsData, wsData] = await Promise.all([
      AppService.GetSkills(),
      AppService.GetWorkspaces()
    ])
    skills.value = skillsData
    workspaces.value = wsData
    await nextTick()
    if (scrollContainer) scrollContainer.scrollTop = scrollTop
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load skills'
  } finally {
    if (isInitialLoad) loading.value = false
  }
}

function isHttpUrl(url: string): boolean {
  return url.startsWith('http://') || url.startsWith('https://')
}

const detailSkill = ref<SkillInfo | null>(null)

function handleDetail(skill: SkillInfo) {
  detailSkill.value = skill
}

const confirmRemoveName = ref<string | null>(null)
const confirmRemoveGroup = ref<SourceGroup | null>(null)

function handleRemove(name: string) {
  confirmRemoveName.value = name
}

async function executeRemove() {
  const name = confirmRemoveName.value
  confirmRemoveName.value = null
  if (!name) return
  try {
    await AppService.RemoveSkillFromWorkspace(name, activeWorkspace.value?.name || 'default')
    await fetchAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to remove skill from workspace'
  }
}

function handleRemoveGroup(group: SourceGroup) {
  confirmRemoveGroup.value = group
}

async function executeRemoveGroup() {
  const group = confirmRemoveGroup.value
  confirmRemoveGroup.value = null
  if (!group) return
  const wsName = activeWorkspace.value?.name || 'default'
  const removable = group.skills.filter(s => !s.isSystem)
  const count = removable.length
  try {
    for (const skill of removable) {
      await AppService.RemoveSkillFromWorkspace(skill.name, wsName)
    }
    await fetchAll()
    const plural = count === 1 ? '' : 's'
    toast.value = {
      message: `Removed ${count} skill${plural} from ${wsName}`,
      type: 'success',
    }
  } catch (e) {
    toast.value = {
      message: e instanceof Error ? e.message : 'Failed to remove skills from workspace',
      type: 'error',
    }
  }
}

// Selection mode
const selectableSkills = computed(() =>
  filteredSkills.value.filter(s => !s.isSystem)
)

const allSelected = computed(() =>
  selectableSkills.value.length > 0 && selectedSkills.size === selectableSkills.value.length
)

function enterSelectionMode() {
  selectionMode.value = true
  selectedSkills.clear()
}

function exitSelectionMode() {
  selectionMode.value = false
  selectedSkills.clear()
}

function toggleSkillSelection(name: string) {
  const skill = filteredSkills.value.find(s => s.name === name)
  if (skill?.isSystem) return
  if (selectedSkills.has(name)) {
    selectedSkills.delete(name)
  } else {
    selectedSkills.add(name)
  }
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedSkills.clear()
  } else {
    for (const skill of selectableSkills.value) {
      selectedSkills.add(skill.name)
    }
  }
}

function hasSelectableSkills(group: SourceGroup): boolean {
  return group.skills.some(s => !s.isSystem)
}

function isGroupAllSelected(group: SourceGroup): boolean {
  const selectable = group.skills.filter(s => !s.isSystem)
  return selectable.length > 0 && selectable.every(s => selectedSkills.has(s.name))
}

function isGroupPartiallySelected(group: SourceGroup): boolean {
  const selectable = group.skills.filter(s => !s.isSystem)
  const selectedCount = selectable.filter(s => selectedSkills.has(s.name)).length
  return selectedCount > 0 && selectedCount < selectable.length
}

function toggleGroupSelection(group: SourceGroup) {
  const selectable = group.skills.filter(s => !s.isSystem)
  if (isGroupAllSelected(group)) {
    for (const skill of selectable) {
      selectedSkills.delete(skill.name)
    }
  } else {
    for (const skill of selectable) {
      selectedSkills.add(skill.name)
    }
  }
}

async function executeBulkRemove() {
  confirmBulkRemove.value = false
  if (selectedSkills.size === 0) return
  bulkRemoving.value = true
  const wsName = activeWorkspace.value?.name || 'default'
  const count = selectedSkills.size
  try {
    for (const skillName of selectedSkills) {
      await AppService.RemoveSkillFromWorkspace(skillName, wsName)
    }
    await fetchAll()
    const plural = count === 1 ? '' : 's'
    toast.value = {
      message: `Removed ${count} skill${plural} from ${wsName}`,
      type: 'success',
    }
    exitSelectionMode()
  } catch (e) {
    toast.value = {
      message: e instanceof Error ? e.message : 'Failed to remove skills from workspace',
      type: 'error',
    }
  } finally {
    bulkRemoving.value = false
  }
}

onMounted(() => {
  fetchAll()
  unsubscribeSkills = Events.On('skills-updated', fetchAll)
  unsubscribeWorkspace = Events.On('workspace-changed', fetchAll)
})

onUnmounted(() => {
  if (unsubscribeSkills) unsubscribeSkills()
  if (unsubscribeWorkspace) unsubscribeWorkspace()
})
</script>

<style scoped>
.skill-list {
  width: 100%;
}

.loading,
.error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 3rem;
  color: var(--text-secondary);
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.error {
  color: var(--danger-color);
}

.skills-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.count {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-secondary {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.btn-secondary:hover {
  background-color: var(--bg-primary);
}

.btn-sm {
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
}

/* Source groups */
.source-group {
  margin-bottom: 1.25rem;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  padding-bottom: 0.375rem;
  border-bottom: 1px solid var(--border-color);
}

.group-checkbox {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
  accent-color: var(--accent-color);
}

.group-badge {
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 600;
  text-transform: uppercase;
}

.group-source {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-primary);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.group-source-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  cursor: pointer;
  text-decoration: none;
  color: var(--text-primary);
  padding: 0.0625rem 0.375rem;
  margin: -0.0625rem -0.375rem;
  border-radius: 4px;
  transition: all 0.15s;
}

.group-source-link:hover {
  color: var(--accent-color);
  background-color: rgba(0, 113, 227, 0.08);
}

.external-icon {
  opacity: 0;
  transition: opacity 0.15s;
  flex-shrink: 0;
}

.group-source-link:hover .external-icon {
  opacity: 0.7;
}

.group-count {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.group-action-btn {
  padding: 0.125rem 0.5rem;
  border: none;
  border-radius: 4px;
  font-size: 0.6875rem;
  font-weight: 500;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  opacity: 0;
  transition: all 0.15s;
}

.group-header:hover .group-action-btn {
  opacity: 1;
}

.group-remove-btn:hover {
  background-color: var(--danger-color);
  color: white;
}

.skills-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

/* Bulk action bar */
.bulk-action-bar {
  position: sticky;
  bottom: 0;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  margin-top: 1rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}

.bulk-count {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
}

.btn-danger {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--danger-color);
  color: white;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-danger:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
