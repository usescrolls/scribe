<template>
  <Teleport to="body">
    <Transition name="modal" appear @after-leave="$emit('close')">
      <div v-if="visible" class="settings-backdrop" @click.self="handleClose">
        <div class="settings-modal">
          <div class="modal-header">
            <h2>Settings</h2>
            <button class="close-btn" @click="handleClose" title="Close">×</button>
          </div>
          <div class="modal-layout">
            <nav class="settings-sidebar">
              <button
                v-for="section in sections"
                :key="section.id"
                :class="['sidebar-item', { active: activeSection === section.id }]"
                @click="activeSection = section.id"
              >
                <span class="sidebar-icon" v-html="section.icon"></span>
                <span class="sidebar-label">{{ section.label }}</span>
              </button>
            </nav>
            <div class="modal-body">
              <section v-if="activeSection === 'agents'" class="settings-section">
                <h3 class="section-title">Agents</h3>
                <AgentStatusPanel />
              </section>

              <section v-else-if="activeSection === 'support'" class="settings-section">
                <h3 class="section-title">Support Scribe</h3>
                <div class="support-section">
                  <p class="support-text">
                    Scribe is free and open source. If you find it useful, consider sponsoring the project to support ongoing development.
                  </p>
                  <button
                    class="sponsor-btn"
                    @click="Browser.OpenURL('https://github.com/sponsors/nunomen')"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
                    </svg>
                    Sponsor on GitHub
                  </button>

                  <div class="social-links">
                    <h4 class="social-heading">Connect</h4>
                    <div class="social-grid">
                      <button
                        class="social-link"
                        title="GitHub"
                        @click="Browser.OpenURL('https://github.com/nunomen')"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"/>
                        </svg>
                        <span>GitHub</span>
                      </button>
                      <button
                        class="social-link"
                        title="X (Twitter)"
                        @click="Browser.OpenURL('https://x.com/1aurentino')"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                        </svg>
                        <span>X</span>
                      </button>
                      <button
                        class="social-link"
                        title="LinkedIn"
                        @click="Browser.OpenURL('https://www.linkedin.com/in/nuno-laurent/')"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                          <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
                        </svg>
                        <span>LinkedIn</span>
                      </button>
                      <button
                        class="social-link"
                        title="Website"
                        @click="Browser.OpenURL('https://nunomen.work')"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <circle cx="12" cy="12" r="10"/>
                          <line x1="2" y1="12" x2="22" y2="12"/>
                          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
                        </svg>
                        <span>Website</span>
                      </button>
                    </div>
                  </div>

                  <p class="made-in">Made with precision in Switzerland</p>
                </div>
              </section>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Browser } from '@wailsio/runtime'
import AgentStatusPanel from './AgentStatusPanel.vue'

defineEmits<{
  close: []
}>()

const visible = ref(true)
const activeSection = ref('agents')

const sections = [
  {
    id: 'agents',
    label: 'Agents',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>',
  },
  {
    id: 'support',
    label: 'Support',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>',
  },
]

function handleClose() {
  visible.value = false
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    handleClose()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.settings-backdrop {
  position: fixed;
  inset: 0;
  z-index: 100;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-modal {
  width: 640px;
  max-width: calc(100vw - 2rem);
  max-height: calc(100vh - 4rem);
  background-color: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.modal-header h2 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}

.close-btn {
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
}

.close-btn:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

/* Sidebar layout */
.modal-layout {
  display: flex;
  flex: 1;
  min-height: 0;
}

.settings-sidebar {
  width: 160px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-color);
  padding: 0.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  background-color: var(--bg-secondary);
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.625rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
  text-align: left;
  width: 100%;
}

.sidebar-item:hover {
  color: var(--text-primary);
  background-color: var(--bg-primary);
}

.sidebar-item.active {
  color: var(--text-primary);
  background-color: var(--bg-primary);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.sidebar-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.sidebar-label {
  white-space: nowrap;
}

.modal-body {
  flex: 1;
  padding: 1.25rem;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.settings-section {
  margin-bottom: 1.5rem;
}

.settings-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 0 0 0.75rem 0;
}

/* Support section */
.support-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.support-text {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}

.sponsor-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.125rem;
  background-color: #bf3989;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.15s;
  align-self: flex-start;
}

.sponsor-btn:hover {
  background-color: #a63179;
}

.social-links {
  display: flex;
  flex-direction: column;
  gap: 0.625rem;
}

.social-heading {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 0;
}

.social-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.social-link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  text-decoration: none;
  font-size: 0.75rem;
  font-weight: 500;
  transition: all 0.15s;
}

.social-link:hover {
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.social-link svg {
  flex-shrink: 0;
}

.made-in {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  margin: 0;
  opacity: 0.7;
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .settings-modal,
.modal-leave-active .settings-modal {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .settings-modal {
  transform: scale(0.95);
  opacity: 0;
}

.modal-leave-to .settings-modal {
  transform: scale(0.95);
  opacity: 0;
}
</style>
