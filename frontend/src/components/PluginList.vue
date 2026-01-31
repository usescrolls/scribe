<template>
  <div class="plugin-list">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>Loading plugins...</span>
    </div>
    <div v-else-if="error" class="error">
      <span>{{ error }}</span>
      <button class="btn-secondary" @click="fetchPlugins">Retry</button>
    </div>
    <EmptyState v-else-if="plugins.length === 0" />
    <div v-else class="plugins">
      <div class="plugins-header">
        <span class="count">{{ plugins.length }} plugin{{ plugins.length !== 1 ? 's' : '' }} installed</span>
      </div>
      <div class="plugins-grid">
        <PluginCard
          v-for="plugin in plugins"
          :key="plugin.name"
          :plugin="plugin"
          @uninstall="handleUninstall"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Dialogs } from '@wailsio/runtime'
import { usePlugins } from '../composables/usePlugins'
import PluginCard from './PluginCard.vue'
import EmptyState from './EmptyState.vue'

const { plugins, loading, error, fetchPlugins, uninstall } = usePlugins()

async function handleUninstall(name: string) {
  console.log('[PluginList] handleUninstall called with:', name)
  try {
    const result = await Dialogs.Question({
      Title: 'Confirm Uninstall',
      Message: `Are you sure you want to uninstall "${name}"?`,
      Buttons: [
        { Label: 'Uninstall', IsDefault: true },
        { Label: 'Cancel', IsCancel: true }
      ]
    })
    console.log('[PluginList] Dialog result:', result)
    if (result === 'Uninstall') {
      console.log('[PluginList] Calling uninstall...')
      const success = await uninstall(name)
      console.log('[PluginList] Uninstall result:', success)
    }
  } catch (err) {
    console.error('[PluginList] Error in handleUninstall:', err)
  }
}
</script>

<style scoped>
.plugin-list {
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

.plugins-header {
  margin-bottom: 1rem;
}

.count {
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.plugins-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
}
</style>
