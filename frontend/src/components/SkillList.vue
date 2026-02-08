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
        </div>
        <div class="skills-list">
          <SkillCard
            v-for="skill in group.skills"
            :key="skill.name"
            :skill="skill"
            :show-remove="true"
            @remove="handleRemove"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Browser, Dialogs, Events } from '@wailsio/runtime'
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

interface SourceGroup {
  source: string
  sourceType: string
  sourceUrl?: string
  skills: SkillInfo[]
}

const groupedSkills = computed<SourceGroup[]>(() => {
  const groups = new Map<string, SourceGroup>()
  for (const skill of filteredSkills.value) {
    const key = skill.source || 'unknown'
    if (!groups.has(key)) {
      groups.set(key, { source: skill.source || 'Unknown source', sourceType: skill.sourceType || 'local', sourceUrl: skill.sourceUrl, skills: [] })
    }
    groups.get(key)!.skills.push(skill)
  }
  return [...groups.values()]
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

.skills-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
</style>
