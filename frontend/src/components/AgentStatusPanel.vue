<template>
  <div class="agent-status-panel" :class="{ collapsed: isCollapsed }">
    <div class="panel-header" @click="toggleCollapse">
      <h3>Agents</h3>
      <span class="summary">{{ installedCount }}/{{ totalCount }} installed</span>
      <span class="collapse-icon">{{ isCollapsed ? '>' : 'v' }}</span>
    </div>
    <div v-if="!isCollapsed" class="agent-list">
      <div
        v-for="agent in sortedAgents"
        :key="agent.id"
        :class="['agent-item', {
          installed: agent.installed,
          selected: selectedAgent === agent.id
        }]"
        @click="handleAgentClick(agent)"
      >
        <div class="agent-icon">
          <span v-if="agent.installed" class="checkmark">&#10003;</span>
          <span v-else class="empty">&#9675;</span>
        </div>
        <div class="agent-info">
          <span class="agent-name">{{ agent.displayName }}</span>
          <span v-if="agent.installed" class="skill-count">
            {{ agent.skillCount }} skill{{ agent.skillCount !== 1 ? 's' : '' }}
          </span>
          <span v-else class="not-installed">Not installed</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAgents } from '../composables/useAgents'

const emit = defineEmits<{
  'agent-selected': [agentId: string | null]
}>()

const { agents, selectedAgent, installedCount, totalCount, selectAgent } = useAgents()

const isCollapsed = ref(false)

const sortedAgents = computed(() => {
  // Show installed agents first, then alphabetically by name
  return [...agents.value].sort((a, b) => {
    if (a.installed !== b.installed) {
      return a.installed ? -1 : 1
    }
    return a.displayName.localeCompare(b.displayName)
  })
})

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function handleAgentClick(agent: { id: string; installed: boolean }) {
  if (!agent.installed) return

  if (selectedAgent.value === agent.id) {
    selectAgent(null)
    emit('agent-selected', null)
  } else {
    selectAgent(agent.id)
    emit('agent-selected', agent.id)
  }
}
</script>

<style scoped>
.agent-status-panel {
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
  cursor: pointer;
  user-select: none;
  border-bottom: 1px solid var(--border-color);
}

.panel-header:hover {
  background-color: var(--bg-primary);
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

.collapse-icon {
  font-size: 0.75rem;
  color: var(--text-secondary);
  width: 1rem;
  text-align: center;
}

.agent-list {
  max-height: 300px;
  overflow-y: auto;
}

.agent-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  transition: background-color 0.15s;
}

.agent-item:hover {
  background-color: var(--bg-primary);
}

.agent-item.selected {
  background-color: rgba(0, 113, 227, 0.1);
}

.agent-item:not(.installed) {
  opacity: 0.5;
  cursor: default;
}

.agent-icon {
  width: 1.25rem;
  height: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
}

.checkmark {
  color: #34c759;
}

.empty {
  color: var(--text-secondary);
}

.agent-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.agent-name {
  font-size: 0.8125rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-count,
.not-installed {
  font-size: 0.6875rem;
  color: var(--text-secondary);
}

.collapsed .panel-header {
  border-bottom: none;
}
</style>
