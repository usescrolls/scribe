<template>
  <div class="step existing-skills-step">
    <h1>Existing Skills Found</h1>

    <div v-if="loading" class="loading">
      <div class="spinner" />
      <p>Scanning for existing skills...</p>
    </div>

    <template v-else-if="skills.length > 0">
      <p class="description">
        We found {{ skills.length }} skill{{ skills.length !== 1 ? 's' : '' }} in your agent directories.
        Would you like to import them into Scribe?
      </p>

      <div class="skills-list">
        <div
          v-for="skill in skills"
          :key="skill.path"
          class="skill-item"
        >
          <div class="skill-info">
            <span class="skill-name">{{ skill.name }}</span>
            <span class="skill-source">{{ skill.agentName }}</span>
          </div>
          <span v-if="skill.isGitRepo" class="git-badge">git</span>
        </div>
      </div>

      <div v-if="conflicts.length > 0" class="conflicts-section">
        <h3>Naming Conflicts</h3>
        <p class="conflict-note">
          Some skills have the same name in different agents. Choose which version to keep:
        </p>
        <div v-for="conflict in conflicts" :key="conflict.name" class="conflict-item">
          <span class="conflict-name">{{ conflict.name }}</span>
          <div class="conflict-options">
            <button
              v-for="source in conflict.sources"
              :key="source.path"
              class="conflict-option"
              @click="$emit('resolve-conflict', source.path)"
            >
              {{ source.agentName }}
            </button>
          </div>
        </div>
      </div>

      <div class="actions">
        <button
          class="primary-button"
          :disabled="loading"
          @click="$emit('import-all')"
        >
          Import All
        </button>
        <button
          class="secondary-button"
          :disabled="loading"
          @click="$emit('delete-all')"
        >
          Delete All & Start Fresh
        </button>
      </div>
    </template>

    <template v-else>
      <p class="no-skills">No existing skills found. Continuing to next step...</p>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { ExistingSkillInfo, SkillConflict } from '../../types/skill'

defineProps<{
  skills: ExistingSkillInfo[]
  conflicts: SkillConflict[]
  loading: boolean
}>()

defineEmits<{
  'import-all': []
  'delete-all': []
  'resolve-conflict': [skillPath: string]
  next: []
}>()
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
  margin-bottom: 1rem;
  color: var(--text-primary);
}

.description {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
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

.skills-list {
  width: 100%;
  max-height: 200px;
  overflow-y: auto;
  margin-bottom: 1.5rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.skill-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
}

.skill-item:last-child {
  border-bottom: none;
}

.skill-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
}

.skill-name {
  font-weight: 500;
  color: var(--text-primary);
}

.skill-source {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.git-badge {
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  padding: 0.125rem 0.375rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-secondary);
}

.conflicts-section {
  width: 100%;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  text-align: left;
}

.conflicts-section h3 {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.conflict-note {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 1rem;
}

.conflict-item {
  margin-bottom: 0.75rem;
}

.conflict-name {
  display: block;
  font-weight: 500;
  margin-bottom: 0.5rem;
}

.conflict-options {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.conflict-option {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.conflict-option:hover {
  border-color: var(--accent-color);
}

.actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  justify-content: center;
}

.primary-button {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 500;
  color: white;
  background-color: var(--accent-color);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.primary-button:hover:not(:disabled) {
  opacity: 0.9;
}

.primary-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.secondary-button {
  padding: 0.75rem 1.5rem;
  font-size: 1rem;
  font-weight: 500;
  color: var(--text-primary);
  background-color: transparent;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s ease;
}

.secondary-button:hover:not(:disabled) {
  border-color: var(--text-secondary);
}

.secondary-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.no-skills {
  color: var(--text-secondary);
  padding: 2rem;
}
</style>
