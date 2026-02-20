<template>
  <div class="step install-demo-step">
    <h1>Install Demo Skill</h1>

    <p class="description">
      We'll install a demo skill to show you how Scribe works.
      This skill will be synced to all your detected agents.
    </p>

    <div class="agents-preview">
      <p class="agents-label">Will be synced to:</p>
      <div class="agents-list">
        <span
          v-for="agent in installedAgents"
          :key="agent.id"
          class="agent-badge"
        >
          {{ agent.displayName }}
        </span>
      </div>
    </div>

    <div class="skill-preview">
      <div class="skill-card">
        <div class="skill-header">
          <span class="skill-name">scribe-welcome</span>
          <span class="skill-type">Demo</span>
        </div>
        <p class="skill-description">
          A welcome skill that introduces Scribe and demonstrates skill formatting
        </p>
      </div>
    </div>

    <div class="button-group">
      <button
        class="primary-button"
        :disabled="installing"
        @click="$emit('install')"
      >
        <span v-if="installing" class="button-content">
          <span class="spinner-small" />
          Installing...
        </span>
        <span v-else>Install Demo Skill</span>
      </button>
      <button
        class="skip-button"
        :disabled="installing"
        @click="$emit('skip')"
      >
        Skip
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AgentStatus } from '../../types/skill'

defineProps<{
  installing: boolean
  installedAgents: AgentStatus[]
}>()

defineEmits<{
  install: []
  skip: []
}>()
</script>

<style scoped>
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 480px;
  width: 100%;
}

h1 {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--text-primary);
}

.description {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.agents-preview {
  width: 100%;
  margin-bottom: 1.5rem;
}

.agents-label {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.agents-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: center;
}

.agent-badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-secondary);
}

.skill-preview {
  width: 100%;
  margin-bottom: 1.5rem;
}

.skill-card {
  padding: 1rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  text-align: left;
}

.skill-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.skill-name {
  font-weight: 600;
  color: var(--text-primary);
}

.skill-type {
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 4px;
}

.skill-description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

.button-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
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
  min-width: 200px;
}

.skip-button {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  background: none;
  border: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.skip-button:hover:not(:disabled) {
  color: var(--text-primary);
}

.skip-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.primary-button:hover:not(:disabled) {
  opacity: 0.9;
}

.primary-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.button-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
