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
      </div>

      <div v-for="group in groupedSkills" :key="group.source" class="source-group">
        <div class="group-header">
          <span class="group-badge">{{ group.sourceType }}</span>
          <a
            v-if="group.sourceUrl"
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
            :show-uninstall="true"
            :show-workspace-picker="true"
            :skill-workspaces="getSkillWorkspaces(skill.name)"
            :all-workspaces="workspaces"
            @add-to-workspace="handleAddToWorkspace"
            @remove-from-workspace="handleRemoveFromWorkspace"
            @uninstall="handleUninstall"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Browser, Events } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import SkillCard from './SkillCard.vue'
import ConfirmDialog from './ConfirmDialog.vue'
import type { SkillInfo, WorkspaceInfo } from '../types/skill'

const allSkillsRaw = ref<SkillInfo[]>([])
const workspaces = ref<WorkspaceInfo[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
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

function isGroupUpdatable(group: SourceGroup): boolean {
  return !!group.sourceType && group.sourceType !== 'local' && group.sourceType !== 'builtin'
}

async function handleUpdateGroup(group: SourceGroup) {
  if (updatingGroup.value) return
  updatingGroup.value = group.source
  try {
    for (const skill of group.skills) {
      await AppService.UpdateSkill(skill.name)
    }
    await fetchAll()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update skills'
  } finally {
    updatingGroup.value = null
  }
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
</style>
