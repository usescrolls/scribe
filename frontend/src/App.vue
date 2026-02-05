<template>
  <OnboardingWizard
    v-if="showOnboarding"
    @complete="onOnboardingComplete"
  />
  <div v-else class="app">
    <header class="header">
      <img src="./assets/icon.png" alt="Scribe" class="app-icon" />
      <h1>Scribe</h1>
      <span class="version">v{{ version }}</span>
    </header>
    <div class="content">
      <aside class="sidebar">
        <SidebarWorkspaceList />
        <AgentStatusPanel @agent-selected="onAgentSelected" />
      </aside>
      <main class="main">
        <div class="main-tabs">
          <button
            :class="['tab', { active: activeTab === 'workspace' }]"
            @click="activeTab = 'workspace'"
          >
            Workspace
          </button>
          <button
            :class="['tab', { active: activeTab === 'browse' }]"
            @click="activeTab = 'browse'; selectedAgent = null"
          >
            Browse All
          </button>
        </div>
        <SkillList v-if="activeTab === 'workspace'" :agent-filter="selectedAgent" />
        <BrowseSkills v-else />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SkillList from './components/SkillList.vue'
import BrowseSkills from './components/BrowseSkills.vue'
import SidebarWorkspaceList from './components/SidebarWorkspaceList.vue'
import AgentStatusPanel from './components/AgentStatusPanel.vue'
import OnboardingWizard from './components/OnboardingWizard.vue'
import { AppService } from './bindings/scribe'

const version = ref('1.0.0')
const selectedAgent = ref<string | null>(null)
const showOnboarding = ref(false)
const onboardingChecked = ref(false)
const activeTab = ref<'workspace' | 'browse'>('workspace')

onMounted(async () => {
  try {
    // Check onboarding status first
    const completed = await AppService.IsOnboardingCompleted()
    showOnboarding.value = !completed
    onboardingChecked.value = true

    // Load version
    version.value = await AppService.GetVersion()
  } catch (e) {
    console.error('Failed to initialize app:', e)
    // If we can't check onboarding, show the main app
    onboardingChecked.value = true
  }
})

function onAgentSelected(agentId: string | null) {
  selectedAgent.value = agentId
}

function onOnboardingComplete() {
  showOnboarding.value = false
}
</script>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  overscroll-behavior: none;
}

.header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  -webkit-app-region: drag;
  flex-shrink: 0;
}

.app-icon {
  width: 24px;
  height: 24px;
  -webkit-app-region: no-drag;
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

.content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.sidebar {
  width: 240px;
  flex-shrink: 0;
  padding: 1rem;
  border-right: 1px solid var(--border-color);
  overflow-y: auto;
  overscroll-behavior: none;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.main {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
  overscroll-behavior: none;
}

.main-tabs {
  display: flex;
  gap: 0.25rem;
  margin-bottom: 1rem;
  padding: 0.25rem;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  width: fit-content;
}

.tab {
  padding: 0.375rem 0.875rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
}

.tab:hover {
  color: var(--text-primary);
}

.tab.active {
  background-color: var(--bg-primary);
  color: var(--text-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* Responsive: hide sidebar on small screens */
@media (max-width: 640px) {
  .sidebar {
    display: none;
  }
}
</style>
