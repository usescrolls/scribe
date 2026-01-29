<template>
  <div class="plugin-card">
    <div class="plugin-header">
      <h3 class="name">{{ plugin.name }}</h3>
      <span v-if="plugin.version" class="version">v{{ plugin.version }}</span>
    </div>
    <p v-if="plugin.description" class="description">{{ plugin.description }}</p>
    <div class="plugin-meta">
      <span class="source-type">{{ plugin.sourceType }}</span>
      <span class="source">{{ plugin.source }}</span>
    </div>
    <div v-if="plugin.author" class="author">
      <span>by {{ plugin.author }}</span>
    </div>
    <div class="plugin-actions">
      <button class="btn-danger" @click="$emit('uninstall', plugin.name)">
        Uninstall
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PluginInfo } from '../types/plugin'

defineProps<{
  plugin: PluginInfo
}>()

defineEmits<{
  uninstall: [name: string]
}>()
</script>

<style scoped>
.plugin-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.plugin-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.name {
  font-size: 1rem;
  font-weight: 600;
}

.version {
  font-size: 0.75rem;
  color: var(--text-secondary);
  padding: 0.125rem 0.375rem;
  background-color: var(--bg-primary);
  border-radius: 4px;
}

.description {
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.4;
}

.plugin-meta {
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

.author {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.plugin-actions {
  margin-top: 0.5rem;
  display: flex;
  justify-content: flex-end;
}
</style>
