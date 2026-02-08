<template>
  <div class="skill-list">
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
      </div>
      <div class="skills-grid">
        <SkillCard
          v-for="skill in filteredSkills"
          :key="skill.name"
          :skill="skill"
          :show-remove="true"
          @remove="handleRemove"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Dialogs, Events } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import SkillCard from './SkillCard.vue'
import EmptyState from './EmptyState.vue'
import type { SkillInfo, WorkspaceInfo } from '../types/skill'

const skills = ref<SkillInfo[]>([])
const workspaces = ref<WorkspaceInfo[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
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

async function fetchAll() {
  try {
    loading.value = true
    error.value = null
    const [skillsData, wsData] = await Promise.all([
      AppService.GetSkills(),
      AppService.GetWorkspaces()
    ])
    skills.value = skillsData
    workspaces.value = wsData
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load skills'
  } finally {
    loading.value = false
  }
}

async function handleRemove(name: string) {
  try {
    const result = await Dialogs.Question({
      Title: 'Remove from Workspace',
      Message: `Remove "${name}" from this workspace?`,
      Buttons: [
        { Label: 'Remove', IsDefault: true },
        { Label: 'Cancel', IsCancel: true }
      ]
    })
    if (result === 'Remove') {
      await AppService.RemoveSkillFromWorkspace(name, activeWorkspace.value?.name || 'default')
      await fetchAll()
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to remove skill from workspace'
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
  margin-bottom: 0.75rem;
}

.count {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.skills-grid {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
</style>
