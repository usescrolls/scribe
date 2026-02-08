<template>
  <div class="browse-skills">
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
          <span class="group-source">{{ group.source }}</span>
          <span class="group-count">{{ group.skills.length }}</span>
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
import { Dialogs, Events } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import SkillCard from './SkillCard.vue'
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
  skills: SkillInfo[]
}

const groupedSkills = computed<SourceGroup[]>(() => {
  const groups = new Map<string, SourceGroup>()
  for (const skill of allSkillsRaw.value) {
    const key = skill.source || 'unknown'
    if (!groups.has(key)) {
      groups.set(key, { source: skill.source || 'Unknown source', sourceType: skill.sourceType || 'local', skills: [] })
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

async function handleUninstall(name: string) {
  try {
    const result = await Dialogs.Question({
      Title: 'Uninstall Skill',
      Message: `Uninstall "${name}"? This will remove it from all workspaces.`,
      Buttons: [
        { Label: 'Uninstall', IsDefault: true },
        { Label: 'Cancel', IsCancel: true }
      ]
    })
    if (result === 'Uninstall') {
      await AppService.RemoveSkill(name)
      await fetchAll()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to uninstall skill'
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

.group-count {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.skills-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
</style>
