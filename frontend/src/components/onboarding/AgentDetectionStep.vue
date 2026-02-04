<template>
  <div class="step agent-detection-step">
    <h1>Detecting Coding Agents</h1>

    <div v-if="loading && agents.length === 0" class="loading">
      <div class="spinner" />
      <p>Scanning for installed agents...</p>
    </div>

    <template v-else>
      <div class="agent-summary">
        <span v-if="hasInstalledAgents" class="found">
          Found {{ installedCount }} coding agent{{ installedCount !== 1 ? 's' : '' }} on your system
        </span>
        <span v-else class="not-found">
          No coding agents detected
        </span>
      </div>

      <div class="agents-list">
        <div
          v-for="agent in agents"
          :key="agent.id"
          class="agent-item"
          :class="{ installed: agent.installed }"
        >
          <span class="agent-icon">{{ agent.installed ? '✓' : '○' }}</span>
          <span class="agent-name">{{ agent.displayName }}</span>
        </div>
      </div>

      <div v-if="!hasInstalledAgents" class="blocked-message">
        <p>
          Please install Claude Code, Cursor, or another supported coding agent to continue.
        </p>
        <p class="scanning-note">
          Scanning every 30 seconds...
        </p>
      </div>

      <button
        v-else
        class="primary-button"
        @click="$emit('next')"
      >
        Continue
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AgentStatus } from '../../types/skill'

const props = defineProps<{
  agents: AgentStatus[]
  loading: boolean
  hasInstalledAgents: boolean
}>()

defineEmits<{
  next: []
}>()

const installedCount = computed(() => props.agents.filter(a => a.installed).length)
</script>

<style scoped>
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 560px;
  width: 100%;
}

h1 {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  color: var(--text-primary);
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  padding: 2rem;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.agent-summary {
  margin-bottom: 1.5rem;
  font-size: 1rem;
}

.found {
  color: var(--text-primary);
}

.not-found {
  color: var(--danger-color);
}

.agents-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 0.5rem;
  width: 100%;
  max-height: 300px;
  overflow-y: auto;
  padding: 0.5rem;
  margin-bottom: 1.5rem;
}

.agent-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  background-color: var(--bg-secondary);
  font-size: 0.875rem;
  opacity: 0.5;
}

.agent-item.installed {
  opacity: 1;
  background-color: var(--bg-secondary);
}

.agent-icon {
  font-size: 0.875rem;
}

.agent-item.installed .agent-icon {
  color: #22c55e;
}

.agent-name {
  color: var(--text-primary);
}

.blocked-message {
  padding: 1rem;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.blocked-message p {
  color: var(--text-secondary);
  margin: 0;
}

.scanning-note {
  font-size: 0.875rem;
  margin-top: 0.5rem !important;
  opacity: 0.7;
}

.primary-button {
  padding: 0.75rem 2rem;
  font-size: 1rem;
  font-weight: 500;
  color: white;
  background-color: var(--accent-color);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.primary-button:hover {
  opacity: 0.9;
}
</style>
