<template>
  <div class="app">
    <header class="header">
      <h1>Scribe</h1>
      <span class="version">v{{ version }}</span>
    </header>
    <main class="main">
      <PluginList />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import PluginList from './components/PluginList.vue'
import { AppService } from './bindings/scribe'

const version = ref('1.0.0')

onMounted(async () => {
  try {
    version.value = await AppService.GetVersion()
  } catch (e) {
    console.error('Failed to get version:', e)
  }
})
</script>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  -webkit-app-region: drag;
}

.header h1 {
  font-size: 1.25rem;
  font-weight: 600;
}

.version {
  font-size: 0.75rem;
  color: var(--text-secondary);
  padding: 0.125rem 0.5rem;
  background-color: var(--bg-primary);
  border-radius: 4px;
}

.main {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
}
</style>
