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
        :disabled="searching"
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
      <div class="results-header">
        <span class="results-count">{{ results.totalCount }} result{{ results.totalCount !== 1 ? 's' : '' }}</span>
        <div v-if="totalPages > 1" class="pagination">
          <button class="page-btn" :disabled="currentPage <= 1 || searching" @click="goToPage(currentPage - 1)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6"></polyline>
            </svg>
          </button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button class="page-btn" :disabled="currentPage >= totalPages || searching" @click="goToPage(currentPage + 1)">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="9 6 15 12 9 18"></polyline>
            </svg>
          </button>
        </div>
      </div>
      <div class="repo-grid">
        <div
          v-for="repo in results.repos"
          :key="repo.fullName"
          class="repo-card"
          @click="handleCardClick($event, repo)"
        >
          <div class="repo-card-header">
            <img v-if="repo.avatarUrl" :src="repo.avatarUrl" class="repo-avatar" alt="" />
            <span class="repo-name">{{ repo.fullName }}</span>
          </div>
          <p v-if="repo.description" class="repo-desc">{{ repo.description }}</p>
          <div class="repo-card-footer">
            <div class="repo-badges">
              <span class="badge badge-stars" title="Stars">
                <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                  <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
                </svg>
                {{ formatStars(repo.stars) }}
              </span>
              <span v-if="repo.skillCount > 0" class="badge badge-skills" title="Skills found">{{ repo.skillCount }} skill{{ repo.skillCount !== 1 ? 's' : '' }}</span>
            </div>
            <div class="repo-actions">
              <button
                class="btn-icon"
                title="Open in browser"
                @click.stop="handleOpenLink(repo)"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
                  <polyline points="15 3 21 3 21 9"></polyline>
                  <line x1="10" y1="14" x2="21" y2="3"></line>
                </svg>
              </button>
              <button
                class="btn-primary"
                @click.stop="handleInstall(repo)"
              >
                Install
              </button>
            </div>
          </div>
        </div>
      </div>
      <!-- Bottom pagination -->
      <div v-if="totalPages > 1" class="pagination pagination-bottom">
        <button class="page-btn" :disabled="currentPage <= 1 || searching" @click="goToPage(currentPage - 1)">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"></polyline>
          </svg>
        </button>
        <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
        <button class="page-btn" :disabled="currentPage >= totalPages || searching" @click="goToPage(currentPage + 1)">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="9 6 15 12 9 18"></polyline>
          </svg>
        </button>
      </div>
    </div>

    <!-- No results -->
    <div v-else-if="results && results.repos.length === 0 && !searching" class="empty-state">
      <p>No skill repositories found{{ lastQuery ? ' for "' : '' }}<strong v-if="lastQuery">{{ lastQuery }}</strong>{{ lastQuery ? '"' : '' }}</p>
    </div>

    <!-- Detail modal -->
    <RepoReadmeModal
      v-if="detailRepo"
      :repo="detailRepo"
      @close="detailRepo = null"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Browser } from '@wailsio/runtime'
import { AppService } from '../bindings/scribe'
import type { MarketplaceResult, MarketplaceRepo } from '../types/skill'
import RepoReadmeModal from './RepoReadmeModal.vue'

const emit = defineEmits<{
  'install-from-source': [source: string]
}>()

const PAGE_SIZE = 30

const query = ref('')
const lastQuery = ref('')
const searching = ref(false)
const error = ref<string | null>(null)
const results = ref<MarketplaceResult | null>(null)
const searchInput = ref<HTMLInputElement | null>(null)
const detailRepo = ref<MarketplaceRepo | null>(null)
const currentPage = ref(1)
const totalPages = computed(() => {
  if (!results.value) return 1
  return Math.max(1, Math.ceil(results.value.totalCount / PAGE_SIZE))
})

onMounted(() => {
  searchInput.value?.focus()
  loadInitial()
})

async function loadInitial() {
  searching.value = true
  error.value = null
  try {
    const res = await AppService.SearchMarketplace('github', '', 1)
    results.value = res as MarketplaceResult
  } catch (e) {
    error.value = extractError(e)
  } finally {
    searching.value = false
  }
}

async function handleSearch() {
  if (searching.value) return

  searching.value = true
  error.value = null
  lastQuery.value = query.value.trim()
  currentPage.value = 1

  try {
    const res = await AppService.SearchMarketplace('github', lastQuery.value, 1)
    results.value = res as MarketplaceResult
  } catch (e) {
    error.value = extractError(e)
    results.value = null
  } finally {
    searching.value = false
  }
}

async function goToPage(page: number) {
  if (searching.value || page < 1 || page > totalPages.value) return

  searching.value = true
  error.value = null

  try {
    const res = await AppService.SearchMarketplace('github', lastQuery.value, page)
    results.value = res as MarketplaceResult
    currentPage.value = page
  } catch (e) {
    error.value = extractError(e)
  } finally {
    searching.value = false
  }
}

function handleInstall(repo: MarketplaceRepo) {
  emit('install-from-source', repo.fullName)
}

function handleOpenLink(repo: MarketplaceRepo) {
  Browser.OpenURL(repo.url)
}

function handleCardClick(_event: MouseEvent, repo: MarketplaceRepo) {
  detailRepo.value = repo
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
  max-width: 100%;
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

.results-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.625rem;
}

.results-count {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.pagination {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.pagination-bottom {
  justify-content: center;
  margin-top: 0.75rem;
}

.page-btn {
  width: 1.5rem;
  height: 1.5rem;
  padding: 0;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.page-btn:hover:not(:disabled) {
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.page-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  min-width: 3rem;
  text-align: center;
}

.repo-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.5rem;
}

.repo-card {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  transition: border-color 0.15s;
  cursor: pointer;
  min-height: 0;
}

.repo-card:hover {
  border-color: var(--accent-color);
}

.repo-card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.repo-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
}

.repo-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.repo-badges {
  display: flex;
  gap: 0.25rem;
  flex-wrap: wrap;
}

.repo-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.375rem;
  margin-top: auto;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.1875rem;
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
  font-size: 0.6875rem;
  color: var(--text-secondary);
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  line-height: 1.4;
}

/* Actions */
.repo-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.btn-icon {
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.btn-icon:hover {
  border-color: var(--accent-color);
  color: var(--accent-color);
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


/* Empty state */
.empty-state {
  margin-top: 2rem;
  text-align: center;
}

.empty-state p {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

</style>
