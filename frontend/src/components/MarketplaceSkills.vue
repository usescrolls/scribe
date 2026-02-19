<template>
  <div class="marketplace">
    <div class="marketplace-header">
      <h2>Marketplace</h2>
      <p class="subtitle">Discover and install skills from GitHub</p>
    </div>

    <!-- Search form -->
    <div class="search-form">
      <input
        ref="searchInput"
        v-model="query"
        type="text"
        placeholder="Search for skills..."
        :disabled="searching"
        @keyup.enter="handleSearch"
      />
      <button
        class="search-btn"
        :disabled="!query.trim() || searching"
        @click="handleSearch"
      >
        <template v-if="searching">
          <div class="spinner-sm"></div>
          Searching...
        </template>
        <template v-else>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          Search
        </template>
      </button>
    </div>

    <!-- Error -->
    <div v-if="error" class="result result-error">
      <div class="result-content">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="15" y1="9" x2="9" y2="15"></line>
          <line x1="9" y1="9" x2="15" y2="15"></line>
        </svg>
        <span>{{ error }}</span>
      </div>
      <button class="dismiss-btn" @click="error = null">&times;</button>
    </div>

    <!-- Results -->
    <div v-if="results && results.repos.length > 0" class="results-section">
      <span class="results-count">{{ results.totalCount }} result{{ results.totalCount !== 1 ? 's' : '' }}</span>
      <div class="repo-list">
        <div
          v-for="repo in results.repos"
          :key="repo.fullName"
          class="repo-card"
        >
          <div class="repo-info">
            <div class="repo-name-row">
              <span class="repo-name">{{ repo.fullName }}</span>
              <div class="repo-badges">
                <span class="badge badge-stars" title="Stars">{{ formatStars(repo.stars) }}</span>
                <span v-if="repo.skillCount > 0" class="badge badge-skills" title="Skills found">{{ repo.skillCount }} skill{{ repo.skillCount !== 1 ? 's' : '' }}</span>
              </div>
            </div>
            <p v-if="repo.description" class="repo-desc">{{ repo.description }}</p>
          </div>
          <div class="repo-actions">
            <button
              v-if="installedRepos.has(repo.fullName)"
              class="btn-installed"
              disabled
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
              Installed
            </button>
            <button
              v-else-if="installingRepo === repo.fullName"
              class="btn-installing"
              disabled
            >
              <div class="spinner-sm"></div>
              Installing...
            </button>
            <button
              v-else
              class="btn-primary"
              @click="handleInstall(repo)"
            >
              Install
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- No results -->
    <div v-else-if="results && results.repos.length === 0 && !searching" class="empty-state">
      <p>No skill repositories found for "<strong>{{ lastQuery }}</strong>"</p>
    </div>

    <!-- Initial state -->
    <div v-else-if="!searching && !error" class="initial-state">
      <p class="initial-hint">Search GitHub for repositories containing agent skills</p>
    </div>

    <!-- Toast -->
    <Transition name="toast">
      <div v-if="toast" class="toast" :class="toast.type">
        {{ toast.message }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { AppService } from '../bindings/scribe'
import type { MarketplaceResult, MarketplaceRepo } from '../types/skill'

const query = ref('')
const lastQuery = ref('')
const searching = ref(false)
const error = ref<string | null>(null)
const results = ref<MarketplaceResult | null>(null)
const searchInput = ref<HTMLInputElement | null>(null)

const installingRepo = ref<string | null>(null)
const installedRepos = reactive(new Set<string>())

const toast = ref<{ message: string; type: 'success' | 'error' } | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => {
  searchInput.value?.focus()
})

async function handleSearch() {
  const q = query.value.trim()
  if (!q || searching.value) return

  searching.value = true
  error.value = null
  lastQuery.value = q

  try {
    const res = await AppService.SearchMarketplace('github', q, 1)
    results.value = res as MarketplaceResult
  } catch (e) {
    error.value = extractError(e)
    results.value = null
  } finally {
    searching.value = false
  }
}

async function handleInstall(repo: MarketplaceRepo) {
  if (installingRepo.value) return

  installingRepo.value = repo.fullName
  try {
    const res = await AppService.InstallFromSource(repo.fullName)
    if (!res) {
      showToast('Installation failed', 'error')
    } else if (res.success) {
      installedRepos.add(repo.fullName)
      showToast(`Installed ${res.skillsCount} skill${res.skillsCount !== 1 ? 's' : ''} from ${repo.fullName}`, 'success')
    } else {
      showToast(res.errorMessage || 'Installation failed', 'error')
    }
  } catch (e) {
    showToast(extractError(e), 'error')
  } finally {
    installingRepo.value = null
  }
}

function showToast(message: string, type: 'success' | 'error') {
  if (toastTimer) clearTimeout(toastTimer)
  toast.value = { message, type }
  toastTimer = setTimeout(() => { toast.value = null }, 4000)
}

function formatStars(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}

function extractError(e: unknown): string {
  let msg: string
  if (e instanceof Error) msg = e.message
  else if (typeof e === 'object' && e !== null && 'message' in e) msg = String((e as { message: unknown }).message)
  else msg = String(e)
  if (msg.startsWith('{')) {
    try {
      const parsed = JSON.parse(msg)
      if (parsed && typeof parsed.message === 'string') return parsed.message
    } catch { /* not JSON */ }
  }
  return msg
}
</script>

<style scoped>
.marketplace {
  max-width: 560px;
  margin: 0 auto;
}

.marketplace-header {
  margin-bottom: 1.5rem;
}

.marketplace-header h2 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.25rem 0;
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin: 0;
}

/* Search form — matches Install tab */
.search-form {
  display: flex;
  gap: 0.5rem;
}

.search-form input {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.875rem;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-family: inherit;
}

.search-form input:focus {
  outline: none;
  border-color: var(--accent-color);
}

.search-form input:disabled {
  opacity: 0.6;
}

.search-btn {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  white-space: nowrap;
  transition: opacity 0.15s;
}

.search-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.search-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner-sm {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Error */
.result {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  margin-top: 0.75rem;
  padding: 0.625rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8125rem;
}

.result-error {
  background-color: rgba(255, 59, 48, 0.1);
  color: var(--danger-color);
}

.result-content {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
}

.result-content svg {
  flex-shrink: 0;
  margin-top: 1px;
}

.dismiss-btn {
  flex-shrink: 0;
  width: 1.25rem;
  height: 1.25rem;
  padding: 0;
  border: none;
  background: transparent;
  color: inherit;
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
  border-radius: 4px;
  opacity: 0.6;
  transition: opacity 0.15s;
}

.dismiss-btn:hover {
  opacity: 1;
}

/* Results */
.results-section {
  margin-top: 1.25rem;
}

.results-count {
  display: block;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.625rem;
}

.repo-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.repo-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  transition: border-color 0.15s;
}

.repo-card:hover {
  border-color: var(--accent-color);
}

.repo-info {
  flex: 1;
  min-width: 0;
}

.repo-name-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.repo-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.repo-badges {
  display: flex;
  gap: 0.25rem;
}

.badge {
  font-size: 0.625rem;
  font-weight: 600;
  padding: 0.0625rem 0.3125rem;
  border-radius: 3px;
  white-space: nowrap;
}

.badge-stars {
  background-color: rgba(255, 204, 0, 0.15);
  color: #b8860b;
}

@media (prefers-color-scheme: dark) {
  .badge-stars {
    background-color: rgba(255, 204, 0, 0.12);
    color: #ffd700;
  }
}

.badge-skills {
  background-color: rgba(0, 113, 227, 0.1);
  color: var(--accent-color);
}

.repo-desc {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin: 0.25rem 0 0 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Actions */
.repo-actions {
  flex-shrink: 0;
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--accent-color);
  color: white;
  cursor: pointer;
  transition: opacity 0.15s;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-installing {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: not-allowed;
}

.btn-installed {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  background-color: rgba(52, 199, 89, 0.1);
  color: var(--success-color);
  cursor: default;
}

/* Empty & initial states */
.empty-state,
.initial-state {
  margin-top: 2rem;
  text-align: center;
}

.empty-state p,
.initial-hint {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

/* Toast */
.toast {
  position: fixed;
  bottom: 1.5rem;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  z-index: 1000;
  pointer-events: none;
}

.toast.success {
  background-color: var(--success-color);
  color: white;
}

.toast.error {
  background-color: var(--danger-color);
  color: white;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(0.5rem);
}
</style>
