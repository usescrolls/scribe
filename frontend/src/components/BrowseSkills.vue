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
    <div v-else-if="allSkills.length === 0" class="empty">
      <p>No skills installed yet.</p>
    </div>
    <div v-else class="skills">
      <div class="skills-header">
        <span class="count">{{ allSkills.length }} skill{{ allSkills.length !== 1 ? 's' : '' }} available</span>
        <span v-if="activeWorkspaceName" class="workspace-hint">
          Adding to: <strong>{{ activeWorkspaceName }}</strong>
        </span>
      </div>
      <div class="skills-list">
        <SkillCard
          v-for="skill in allSkills"
          :key="skill.name"
          :skill="skill"
          :show-uninstall="true"
          :show-add="!isInWorkspace(skill.name)"
          @add="handleAdd"
          @uninstall="handleUninstall"
        />
      </div>
      <div v-if="skillsInWorkspace.length > 0" class="in-workspace-section">
        <span class="section-label">Already in workspace ({{ skillsInWorkspace.length }})</span>
        <div class="skills-list muted">
          <SkillCard
            v-for="skill in skillsInWorkspace"
            :key="skill.name"
            :skill="skill"
            :show-uninstall="true"
            :show-add="false"
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

const activeWorkspace = computed(() => {
  return workspaces.value.find(ws => ws.isActive)
})

const activeWorkspaceName = computed(() => {
  return activeWorkspace.value?.name || 'default'
})

const workspaceSkillNames = computed(() => {
  return new Set(activeWorkspace.value?.skills || [])
})

const skillsNotInWorkspace = computed(() => {
  return allSkillsRaw.value.filter(skill => !workspaceSkillNames.value.has(skill.name))
})

const skillsInWorkspace = computed(() => {
  return allSkillsRaw.value.filter(skill => workspaceSkillNames.value.has(skill.name))
})

const allSkills = computed(() => {
  return skillsNotInWorkspace.value
})

function isInWorkspace(skillName: string): boolean {
  return workspaceSkillNames.value.has(skillName)
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

async function handleAdd(skillName: string) {
  try {
    await AppService.AddSkillToWorkspace(skillName, activeWorkspaceName.value)
    await fetchAll()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to add skill'
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

.workspace-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.workspace-hint strong {
  color: var(--accent-color);
}

.skills-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.skills-list.muted {
  opacity: 0.6;
}

.in-workspace-section {
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

.section-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}
</style>
