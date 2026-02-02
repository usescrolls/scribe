<template>
  <div class="workspace-selector">
    <select
      v-model="selectedWorkspace"
      @change="handleChange"
      :disabled="loading"
    >
      <option
        v-for="ws in workspaces"
        :key="ws.name"
        :value="ws.name"
      >
        {{ ws.name }} ({{ ws.skills.length }} skill{{ ws.skills.length !== 1 ? 's' : '' }})
      </option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useWorkspaces } from '../composables/useWorkspaces'

const { workspaces, activeWorkspace, loading, switchWorkspace } = useWorkspaces()

const selectedWorkspace = ref(activeWorkspace.value)

// Keep local selection in sync with actual active workspace
watch(activeWorkspace, (newValue) => {
  selectedWorkspace.value = newValue
})

async function handleChange() {
  if (selectedWorkspace.value !== activeWorkspace.value) {
    await switchWorkspace(selectedWorkspace.value)
  }
}
</script>

<style scoped>
.workspace-selector {
  display: flex;
  align-items: center;
}

select {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  min-width: 140px;
  -webkit-app-region: no-drag;
}

select:hover {
  border-color: var(--accent-color);
}

select:focus {
  outline: none;
  border-color: var(--accent-color);
  box-shadow: 0 0 0 2px rgba(0, 113, 227, 0.2);
}

select:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
