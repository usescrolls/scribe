<template>
  <div class="skill-card">
    <div class="skill-header">
      <h3 class="name">{{ skill.name }}</h3>
    </div>
    <p v-if="skill.description" class="description">{{ skill.description }}</p>
    <div class="skill-meta">
      <span class="source-type">{{ skill.sourceType }}</span>
      <span class="source">{{ skill.source }}</span>
    </div>
    <div v-if="skill.agents.length > 0" class="agents">
      <span
        v-for="agent in skill.agents"
        :key="agent"
        class="agent-badge"
      >
        {{ formatAgentName(agent) }}
      </span>
    </div>
    <div class="skill-actions">
      <button class="btn-danger" @click="$emit('uninstall', skill.name)">
        Uninstall
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SkillInfo } from '../types/skill'

defineProps<{
  skill: SkillInfo
}>()

defineEmits<{
  uninstall: [name: string]
}>()

function formatAgentName(agentId: string): string {
  // Convert agent ID to display name (e.g., "claude-code" -> "Claude Code")
  return agentId
    .split('-')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}
</script>

<style scoped>
.skill-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.skill-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.name {
  font-size: 1rem;
  font-weight: 600;
}

.description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.skill-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
}

.source-type {
  padding: 0.125rem 0.5rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 4px;
  text-transform: uppercase;
  font-weight: 500;
}

.source {
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.agents {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.agent-badge {
  font-size: 0.6875rem;
  padding: 0.125rem 0.5rem;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-secondary);
}

.skill-actions {
  margin-top: 0.5rem;
  display: flex;
  justify-content: flex-end;
}
</style>
