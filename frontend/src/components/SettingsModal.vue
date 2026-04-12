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

              <section v-else-if="activeSection === 'updates'" class="settings-section">
                <h3 class="section-title">Updates</h3>
                <div class="updates-section">
                  <div class="version-info">
                    <div class="version-row">
                      <span class="version-label">Current version</span>
                      <span class="version-value">{{ updateInfo?.currentVersion ?? '...' }}</span>
                    </div>
                    <div class="version-row">
                      <span class="version-label">Latest version</span>
                      <span class="version-value">{{ updateInfo?.latestVersion ?? '...' }}</span>
                    </div>
                    <div v-if="upgradeSuccess" class="update-available">
                      <span class="upgrade-success">Upgraded successfully!</span>
                      <button class="upgrade-btn" @click="restartApp">
                        Restart Scribe
                      </button>
                    </div>
                    <div v-else-if="updateInfo?.updateAvailable" class="update-available">
                      <span class="update-badge">Update available</span>

                      <template v-if="installMethod === 'binary' || installMethod === 'app-bundle'">
                        <button
                          class="upgrade-btn"
                          :disabled="upgrading"
                          @click="upgradeApp"
                        >
                          {{ upgrading ? 'Upgrading...' : 'Upgrade now' }}
                        </button>
                        <span v-if="upgradeError" class="upgrade-error">
                          {{ upgradeError }}
                        </span>
                      </template>

                      <template v-else-if="installMethod === 'homebrew'">
                        <span class="homebrew-hint">
                          Run <code>brew upgrade usescrolls/tap/scribe</code> to update
                        </span>
                      </template>

                      <template v-else>
                        <button class="update-link-btn" @click="openReleasePage">
                          View release notes
                        </button>
                      </template>
                    </div>
                    <div v-else-if="updateInfo && !updateInfo.updateAvailable" class="up-to-date">
                      You're running the latest version.
                    </div>
                  </div>

                  <button
                    class="check-update-btn"
                    :disabled="updateLoading"
                    @click="checkForUpdate"
                  >
                    {{ updateLoading ? 'Checking...' : 'Check for Updates' }}
                  </button>

                  <label class="suppress-checkbox">
                    <input
                      type="checkbox"
                      :checked="notificationsDisabled"
                      @change="setNotificationsDisabled(($event.target as HTMLInputElement).checked)"
                    />
                    <span>Don't show update notifications</span>
                  </label>
                </div>
              </section>

              <section v-else-if="activeSection === 'terms'" class="settings-section">
                <h3 class="section-title">Terms & Conditions</h3>
                <div class="terms-section">
                  <div class="terms-box">
                    <div class="terms-content">
                      <template v-for="(clause, i) in termsClauses" :key="i">
                        <h4>{{ i + 1 }}. {{ clause.title }}</h4>
                        <p>{{ clause.body }}</p>
                      </template>
                    </div>
                  </div>
                  <div v-if="termsAcceptedAt" class="terms-accepted-note">
                    Accepted on {{ formatDate(termsAcceptedAt) }}
                  </div>
                </div>
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

              <section v-else-if="activeSection === 'report'" class="settings-section">
                <h3 class="section-title">Report an Issue</h3>
                <div class="report-section">
                  <p class="report-text">
                    Found a bug or have a feature request? Open an issue on GitLab and we'll look into it.
                    Please note that we do not accept pull requests.
                  </p>
                  <button
                    class="report-btn"
                    @click="Browser.OpenURL('https://gitlab.com/usescrolls/scribe/-/issues/new')"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <circle cx="12" cy="12" r="10"/>
                      <line x1="12" y1="8" x2="12" y2="12"/>
                      <line x1="12" y1="16" x2="12.01" y2="16"/>
                    </svg>
                    Open Issue on GitLab
                  </button>
                  <div class="log-hint">
                    <h4 class="log-hint-heading">Include logs if needed</h4>
                    <p class="log-hint-text">
                      If you're reporting a bug, attaching the log file can help us diagnose the problem faster. You can find it at:
                    </p>
                    <code class="log-path">~/.scribe/scribe.log</code>
                  </div>
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
import { useUpdateChecker } from '../composables/useUpdateChecker'
import { AppService } from '../bindings/scribe'

defineEmits<{
  close: []
}>()

const visible = ref(true)
const activeSection = ref('agents')
const termsAcceptedAt = ref('')
const termsClauses = ref<{ title: string; body: string }[]>([])

const {
  updateInfo,
  loading: updateLoading,
  notificationsDisabled,
  installMethod,
  upgrading,
  upgradeError,
  upgradeSuccess,
  checkForUpdate,
  setNotificationsDisabled,
  openReleasePage,
  upgradeApp,
  restartApp,
} = useUpdateChecker()

const sections = [
  {
    id: 'agents',
    label: 'Agents',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>',
  },
  {
    id: 'updates',
    label: 'Updates',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>',
  },
  {
    id: 'terms',
    label: 'Terms',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>',
  },
  {
    id: 'support',
    label: 'Support',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>',
  },
  {
    id: 'report',
    label: 'Report Issue',
    icon: '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>',
  },
]

function formatDate(isoString: string): string {
  try {
    return new Date(isoString).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
  } catch {
    return isoString
  }
}

function handleClose() {
  visible.value = false
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    handleClose()
  }
}

onMounted(async () => {
  document.addEventListener('keydown', handleKeydown)
  termsAcceptedAt.value = await AppService.GetTermsAcceptedAt()
  termsClauses.value = await AppService.GetTermsClauses()
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
  height: 420px;
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

/* Report issue section */
.report-section {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.report-text {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.6;
  margin: 0;
}

.report-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.125rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  border-radius: 8px;
  font-size: 0.8125rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  align-self: flex-start;
}

.report-btn:hover {
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.log-hint {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.75rem 1rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
}

.log-hint-heading {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.log-hint-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0;
}

.log-path {
  font-size: 0.75rem;
  background-color: var(--bg-primary);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: var(--accent-color);
  align-self: flex-start;
}

/* Terms section */
.terms-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.terms-box {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-secondary);
}

.terms-content {
  padding: 1rem;
}

.terms-content h4 {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.25rem 0;
}

.terms-content h4:not(:first-child) {
  margin-top: 0.75rem;
}

.terms-content p {
  font-size: 0.75rem;
  line-height: 1.6;
  color: var(--text-secondary);
  margin: 0;
}

.terms-accepted-note {
  font-size: 0.75rem;
  color: var(--text-secondary);
  opacity: 0.7;
}

/* Updates section */
.updates-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.version-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.version-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8125rem;
}

.version-label {
  color: var(--text-secondary);
}

.version-value {
  font-weight: 600;
  color: var(--text-primary);
}

.update-available {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.25rem;
}

.update-badge {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--accent-color);
  background-color: var(--bg-secondary);
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
}

.update-link-btn {
  font-size: 0.75rem;
  color: var(--accent-color);
  background: none;
  border: none;
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
}

.update-link-btn:hover {
  opacity: 0.8;
}

.upgrade-btn {
  padding: 0.25rem 0.75rem;
  background-color: var(--accent-color);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
}

.upgrade-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.upgrade-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.upgrade-success {
  font-size: 0.75rem;
  color: #22c55e;
  font-weight: 500;
}

.upgrade-error {
  font-size: 0.75rem;
  color: #ef4444;
  font-weight: 500;
}

.homebrew-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.homebrew-hint code {
  background-color: var(--bg-secondary);
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-size: 0.6875rem;
}

.up-to-date {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

.check-update-btn {
  align-self: flex-start;
  padding: 0.5rem 1rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}

.check-update-btn:hover:not(:disabled) {
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.check-update-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.suppress-checkbox {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  cursor: pointer;
}

.suppress-checkbox input[type="checkbox"] {
  accent-color: var(--accent-color);
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
