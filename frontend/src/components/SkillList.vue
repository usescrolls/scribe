<template>
  <div class="skill-list">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>Loading skills...</span>
    </div>
    <div v-else-if="error" class="error">
      <span>{{ error }}</span>
      <button class="btn-secondary" @click="fetchSkills">Retry</button>
    </div>
    <EmptyState v-else-if="filteredSkills.length === 0" />
    <div v-else class="skills">
      <div class="skills-header">
        <span class="count">{{ filteredSkills.length }} skill{{ filteredSkills.length !== 1 ? 's' : '' }} installed</span>
      </div>
      <div class="skills-grid">
        <SkillCard
          v-for="skill in filteredSkills"
          :key="skill.name"
          :skill="skill"
          @uninstall="handleUninstall"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Dialogs } from '@wailsio/runtime'
import { useSkills } from '../composables/useSkills'
import SkillCard from './SkillCard.vue'
import EmptyState from './EmptyState.vue'

const props = defineProps<{
  agentFilter?: string | null
}>()

const { skills, loading, error, fetchSkills, uninstall } = useSkills()

const filteredSkills = computed(() => {
  if (!props.agentFilter) {
    return skills.value
  }
  return skills.value.filter(skill =>
    skill.agents.includes(props.agentFilter!)
  )
})

async function handleUninstall(name: string) {
  console.log('[SkillList] handleUninstall called with:', name)
  try {
    const result = await Dialogs.Question({
      Title: 'Confirm Uninstall',
      Message: `Are you sure you want to uninstall "${name}"?`,
      Buttons: [
        { Label: 'Uninstall', IsDefault: true },
        { Label: 'Cancel', IsCancel: true }
      ]
    })
    console.log('[SkillList] Dialog result:', result)
    if (result === 'Uninstall') {
      console.log('[SkillList] Calling uninstall...')
      const success = await uninstall(name)
      console.log('[SkillList] Uninstall result:', success)
    }
  } catch (err) {
    console.error('[SkillList] Error in handleUninstall:', err)
  }
}
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
  margin-bottom: 1rem;
}

.count {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.skills-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
}
</style>
