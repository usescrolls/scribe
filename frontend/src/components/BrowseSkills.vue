<template>
  <div class="browse-skills">
    <ConfirmDialog
      v-if="confirmUninstallName"
      title="Uninstall Skill"
      :message="`Uninstall &quot;${confirmUninstallName}&quot;? This will remove it from all workspaces.`"
      confirm-label="Uninstall"
      :danger="true"
      @confirm="executeUninstall"
      @cancel="confirmUninstallName = null"
    />
    <ConfirmDialog
      v-if="confirmUninstallGroup"
      title="Uninstall All Skills"
      :message="`Uninstall all ${confirmUninstallGroup.skills.length} skills from &quot;${confirmUninstallGroup.source}&quot;? This will remove them from all workspaces.`"
      confirm-label="Uninstall All"
      :danger="true"
      @confirm="executeUninstallGroup"
      @cancel="confirmUninstallGroup = null"
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
    <div v-else-if="allSkillsRaw.length === 0" class="empty">
      <p>No skills installed yet.</p>
    </div>
    <div v-else class="skills">
      <div class="skills-header">
        <span class="count">{{ allSkillsRaw.length }} skill{{ allSkillsRaw.length !== 1 ? 's' : '' }} installed</span>
        <div class="header-actions">
          <template v-if="selectionMode">
            <button class="btn-secondary btn-sm" @click="toggleSelectAll">
              {{ allSelected ? 'Deselect All' : 'Select All' }}
            </button>
            <button class="btn-secondary btn-sm" @click="exitSelectionMode">Cancel</button>
          </template>
          <button v-else class="btn-secondary btn-sm" @click="enterSelectionMode">Select</button>
        </div>
      </div>

      <div v-for="group in groupedSkills" :key="group.source" class="source-group">
        <div class="group-header">
          <span class="group-badge">{{ group.sourceType }}</span>
          <svg v-if="isGroupPrivate(group)" class="private-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" title="Private repository">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
          </svg>
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
          <span v-if="getGroupVersion(group)" class="group-version" :title="getGroupVersionTooltip(group)">{{ getGroupVersion(group) }}</span>
          <span class="group-count">{{ group.skills.length }}</span>
          <button
            v-if="isGroupUpdatable(group)"
            class="group-action-btn group-update-btn"
            :disabled="updatingGroup === group.source"
            @click="handleUpdateGroup(group)"
          >
            {{ updatingGroup === group.source ? 'Updating...' : 'Update' }}
          </button>
          <button
            v-if="group.skills.length > 1"
            class="group-action-btn group-uninstall-btn"
            @click="handleUninstallGroup(group)"
          >
            Uninstall all
          </button>
        </div>
        <div class="skills-list">
          <SkillCard
            v-for="skill in group.skills"
            :key="skill.name"
            :skill="skill"
            :show-uninstall="!selectionMode"
            :show-workspace-picker="!selectionMode"
            :selectable="selectionMode"
            :selected="selectedSkills.has(skill.name)"
            :skill-workspaces="getSkillWorkspaces(skill.name)"
            :all-workspaces="workspaces"
            @detail="handleDetail"
            @toggle-select="toggleSkillSelection"
            @add-to-workspace="handleAddToWorkspace"
            @remove-from-workspace="handleRemoveFromWorkspace"
            @uninstall="handleUninstall"
          />
        </div>
      </div>

      <!-- Bulk action bar -->
      <div v-if="selectionMode && selectedSkills.size > 0" class="bulk-action-bar">
        <span class="bulk-count">{{ selectedSkills.size }} skill{{ selectedSkills.size !== 1 ? 's' : '' }} selected</span>
        <select v-model="bulkWorkspace" class="bulk-workspace-select">
          <option value="" disabled>Choose workspace...</option>
          <option v-for="ws in workspaces" :key="ws.name" :value="ws.name">{{ ws.name }}</option>
        </select>
        <button
          class="btn-primary btn-sm"
          :disabled="!bulkWorkspace || bulkAdding"
          @click="bulkAddToWorkspace"
        >
          {{ bulkAdding ? 'Adding...' : 'Add to workspace' }}
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
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { Browser, Events } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import SkillCard from './SkillCard.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import ToastNotification from './ToastNotification.vue'
import SkillDetailModal from './SkillDetailModal.vue'
import type { SkillInfo, WorkspaceInfo, UpdateResult } from '../types/skill'

const allSkillsRaw = ref<SkillInfo[]>([])
const workspaces = ref<WorkspaceInfo[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const toast = ref<{ message: string; type: 'success' | 'error' | 'info' } | null>(null)
const selectedSkills = reactive(new Set<string>())
const selectionMode = ref(false)
const bulkWorkspace = ref('')
const bulkAdding = ref(false)
let unsubscribeSkills: { (): void } | null = null
let unsubscribeWorkspace: { (): void } | null = null

interface SourceGroup {
  source: string
  sourceType: string
  sourceUrl?: string
  skills: SkillInfo[]
}

const groupedSkills = computed<SourceGroup[]>(() => {
  const groups = new Map<string, SourceGroup>()
  for (const skill of allSkillsRaw.value) {
    const key = skill.source || 'unknown'
    if (!groups.has(key)) {
      groups.set(key, { source: skill.source || 'Unknown source', sourceType: skill.sourceType || 'local', sourceUrl: skill.sourceUrl, skills: [] })
    }
    groups.get(key)!.skills.push(skill)
  }
  return [...groups.values()]
})

function getSkillWorkspaces(skillName: string): string[] {
  return workspaces.value
    .filter(ws => ws.skills.includes(skillName))
    .map(ws => ws.name)
}

async function fetchAll() {
  try {
    loading.value = true
    error.value = null
    const [skills, ws] = await Promise.all([
      AppService.GetSkills(),
      AppService.GetWorkspaces()
    ])
    allSkillsRaw.value = skills
    workspaces.value = ws
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load skills'
  } finally {
    loading.value = false
  }
}

async function handleAddToWorkspace(skillName: string, workspaceName: string) {
  try {
    await AppService.AddSkillToWorkspace(skillName, workspaceName)
    await fetchAll()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to add skill to workspace'
  }
}

async function handleRemoveFromWorkspace(skillName: string, workspaceName: string) {
  try {
    await AppService.RemoveSkillFromWorkspace(skillName, workspaceName)
    await fetchAll()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to remove skill from workspace'
  }
}

const updatingGroup = ref<string | null>(null)

function isAuthError(msg: string): boolean {
  const lower = msg.toLowerCase()
  return ['authentication required', 'authentication failed', 'permission denied',
    'repository not found', 'access denied', '403', '401'].some(p => lower.includes(p))
}

function isGroupPrivate(group: SourceGroup): boolean {
  return group.skills.some(s => s.isPrivate)
}

function isGroupUpdatable(group: SourceGroup): boolean {
  return !!group.sourceType && group.sourceType !== 'local' && group.sourceType !== 'builtin'
}

async function handleUpdateGroup(group: SourceGroup) {
  if (updatingGroup.value) return
  updatingGroup.value = group.source
  try {
    const results: UpdateResult[] = []
    for (const skill of group.skills) {
      const result = await AppService.UpdateSkill(skill.name)
      if (result) results.push(result as UpdateResult)
    }
    await fetchAll()

    // Show toast with update summary
    const updated = results.filter(r => r.updated)
    if (updated.length > 0) {
      const hashInfo = updated[0].newHash
        ? updated[0].oldHash
          ? ` (${updated[0].oldHash} \u2192 ${updated[0].newHash})`
          : ` (${updated[0].newHash})`
        : ''
      toast.value = {
        message: `Updated ${updated.length} skill${updated.length !== 1 ? 's' : ''}${hashInfo}`,
        type: 'success',
      }
    } else {
      toast.value = { message: 'All skills already up to date', type: 'info' }
    }
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'Failed to update skills'
    const hint = isAuthError(msg)
      ? '. For private repos, ensure git credentials are configured (e.g. gh auth login)'
      : ''
    toast.value = {
      message: msg + hint,
      type: 'error',
    }
  } finally {
    updatingGroup.value = null
  }
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}d ago`
  const months = Math.floor(days / 30)
  return `${months}mo ago`
}

function getGroupVersion(group: SourceGroup): string | null {
  const first = group.skills[0]
  if (!first?.commitHash) return null
  const hash = first.commitHash
  const date = first.commitDate || first.updatedAt
  if (date) return `${hash} · ${relativeTime(date)}`
  return hash
}

function getGroupVersionTooltip(group: SourceGroup): string {
  const first = group.skills[0]
  const parts: string[] = []
  if (first?.commitHash) parts.push(`Commit: ${first.commitHash}`)
  if (first?.commitDate) parts.push(`Date: ${new Date(first.commitDate).toLocaleDateString()}`)
  if (first?.updatedAt) parts.push(`Updated: ${new Date(first.updatedAt).toLocaleDateString()}`)
  return parts.join('\n')
}

function isHttpUrl(url: string): boolean {
  return url.startsWith('http://') || url.startsWith('https://')
}

const detailSkill = ref<SkillInfo | null>(null)

function handleDetail(skill: SkillInfo) {
  detailSkill.value = skill
}

const confirmUninstallName = ref<string | null>(null)
const confirmUninstallGroup = ref<SourceGroup | null>(null)

function handleUninstall(name: string) {
  confirmUninstallName.value = name
}

async function executeUninstall() {
  const name = confirmUninstallName.value
  confirmUninstallName.value = null
  if (!name) return
  try {
    await AppService.RemoveSkill(name)
    await fetchAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to uninstall skill'
  }
}

function handleUninstallGroup(group: SourceGroup) {
  confirmUninstallGroup.value = group
}

async function executeUninstallGroup() {
  const group = confirmUninstallGroup.value
  confirmUninstallGroup.value = null
  if (!group) return
  try {
    for (const skill of group.skills) {
      await AppService.RemoveSkill(skill.name)
    }
    await fetchAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to uninstall skills'
  }
}

function enterSelectionMode() {
  selectionMode.value = true
  selectedSkills.clear()
}

function exitSelectionMode() {
  selectionMode.value = false
  selectedSkills.clear()
  bulkWorkspace.value = ''
}

function toggleSkillSelection(name: string) {
  if (selectedSkills.has(name)) {
    selectedSkills.delete(name)
  } else {
    selectedSkills.add(name)
  }
}

const allSelected = computed(() =>
  allSkillsRaw.value.length > 0 && selectedSkills.size === allSkillsRaw.value.length
)

function toggleSelectAll() {
  if (allSelected.value) {
    selectedSkills.clear()
  } else {
    for (const skill of allSkillsRaw.value) {
      selectedSkills.add(skill.name)
    }
  }
}

async function bulkAddToWorkspace() {
  if (!bulkWorkspace.value || selectedSkills.size === 0) return
  bulkAdding.value = true
  // Temporarily unsub from workspace-changed to avoid N refetches
  if (unsubscribeWorkspace) unsubscribeWorkspace()
  try {
    for (const skillName of selectedSkills) {
      await AppService.AddSkillToWorkspace(skillName, bulkWorkspace.value)
    }
    await fetchAll()
    const count = selectedSkills.size
    toast.value = {
      message: `Added ${count} skill${count !== 1 ? 's' : ''} to ${bulkWorkspace.value}`,
      type: 'success',
    }
    exitSelectionMode()
  } catch (e) {
    toast.value = {
      message: e instanceof Error ? e.message : 'Failed to add skills to workspace',
      type: 'error',
    }
  } finally {
    bulkAdding.value = false
    // Re-subscribe
    unsubscribeWorkspace = Events.On('workspace-changed', fetchAll)
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
.browse-skills {
  width: 100%;
}

.loading,
.error,
.empty {
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

.skills-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.count {
  font-size: 0.875rem;
  color: var(--text-secondary);
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

.group-badge {
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 3px;
  font-size: 0.5625rem;
  font-weight: 600;
  text-transform: uppercase;
}

.private-icon {
  color: var(--text-secondary);
  flex-shrink: 0;
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

.group-version {
  font-size: 0.625rem;
  color: var(--text-secondary);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  background-color: var(--bg-primary);
  padding: 0.0625rem 0.3125rem;
  border-radius: 3px;
  cursor: default;
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

.group-action-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

.group-update-btn:hover:not(:disabled) {
  background-color: var(--accent-color);
  color: white;
}

.group-uninstall-btn:hover {
  background-color: var(--danger-color);
  color: white;
}

.skills-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

/* Header actions */
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-sm {
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
}

.btn-primary {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: all 0.15s;
}

.btn-primary:hover:not(:disabled) {
  filter: brightness(1.1);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: default;
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

.bulk-workspace-select {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.8125rem;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  min-width: 140px;
}
</style>
