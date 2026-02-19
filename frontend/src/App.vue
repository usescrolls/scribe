<template>
  <OnboardingWizard
    v-if="showOnboarding"
    @complete="onOnboardingComplete"
  />
  <div v-else class="app">
    <header class="header">
      <div class="header-spacer"></div>
      <div class="header-center">
        <img src="./assets/icon.png" alt="Scribe" class="app-icon" />
        <h1>Scribe</h1>
      </div>
      <div class="header-right">
        <span class="version">v{{ version }}</span>
        <button class="settings-btn" @click="showSettings = true" title="Settings">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"></circle>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
          </svg>
        </button>
      </div>
    </header>
    <nav class="tab-bar">
      <div class="tab-group">
        <button
          :class="['tab', { active: activeTab === 'workspace' }]"
          @click="activeTab = 'workspace'"
        >
          Workspace
        </button>
        <button
          :class="['tab', { active: activeTab === 'browse' }]"
          @click="activeTab = 'browse'"
        >
          Browse All
        </button>
        <button
          :class="['tab', { active: activeTab === 'install' }]"
          @click="activeTab = 'install'"
        >
          Install
        </button>
        <button
          :class="['tab', { active: activeTab === 'marketplace' }]"
          @click="activeTab = 'marketplace'"
        >
          Marketplace
        </button>
      </div>
      <WorkspaceDropdown />
    </nav>
    <main class="main">
      <Transition name="fade" mode="out-in">
        <SkillList v-if="activeTab === 'workspace'" key="workspace" />
        <BrowseSkills v-else-if="activeTab === 'browse'" key="browse" />
        <InstallSkills v-else-if="activeTab === 'install'" key="install" :initial-source="pendingInstallSource" @consumed="pendingInstallSource = null" />
        <MarketplaceSkills v-else-if="activeTab === 'marketplace'" key="marketplace" @install-from-source="handleMarketplaceInstall" />
      </Transition>
    </main>
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import SkillList from './components/SkillList.vue'
import BrowseSkills from './components/BrowseSkills.vue'
import InstallSkills from './components/InstallSkills.vue'
import MarketplaceSkills from './components/MarketplaceSkills.vue'
import WorkspaceDropdown from './components/WorkspaceDropdown.vue'
import SettingsModal from './components/SettingsModal.vue'
import OnboardingWizard from './components/OnboardingWizard.vue'
import { AppService } from './bindings/scribe'

const version = ref('1.0.0')
const showOnboarding = ref(false)
const onboardingChecked = ref(false)
const showSettings = ref(false)
const activeTab = ref<'workspace' | 'browse' | 'install' | 'marketplace'>('workspace')
const pendingInstallSource = ref<string | null>(null)

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

function onOnboardingComplete() {
  showOnboarding.value = false
}

function handleMarketplaceInstall(source: string) {
  pendingInstallSource.value = source
  activeTab.value = 'install'
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
  justify-content: space-between;
  padding: 0.75rem 1.25rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  -webkit-app-region: drag;
  flex-shrink: 0;
}

.header-spacer,
.header-right {
  flex: 1;
  min-width: 0;
}

.header-spacer {
  pointer-events: none;
}

.header-center {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  flex-shrink: 0;
}

.header-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.625rem;
  -webkit-app-region: no-drag;
}

.app-icon {
  width: 22px;
  height: 22px;
  -webkit-app-region: no-drag;
}

.header h1 {
  font-size: 1.125rem;
  font-weight: 600;
}

.version {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  padding: 0.125rem 0.4375rem;
  background-color: var(--bg-primary);
  border-radius: 4px;
}

.settings-btn {
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  -webkit-app-region: no-drag;
}

.settings-btn:hover {
  background-color: var(--bg-primary);
  color: var(--text-primary);
}

.tab-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1.25rem;
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.tab-group {
  display: flex;
  gap: 0.25rem;
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

.main {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
  overscroll-behavior: none;
}

/* Fade transition (matches onboarding) */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
