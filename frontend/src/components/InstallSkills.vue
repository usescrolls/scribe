<template>
  <div class="install-skills">
    <div class="install-header">
      <h2>Install Skills</h2>
      <p class="subtitle">Add skills from GitHub, GitLab, or a zip URL</p>
    </div>

    <div class="install-form">
      <input
        ref="sourceInput"
        v-model="sourceStr"
        type="text"
        placeholder="owner/repo, GitHub URL, or zip URL"
        :disabled="installing"
        @keyup.enter="handleInstall"
      />
      <button
        class="install-btn"
        :disabled="!sourceStr.trim() || installing"
        @click="handleInstall"
      >
        <template v-if="installing">
          <div class="spinner-sm"></div>
          Installing...
        </template>
        <template v-else>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
            <polyline points="7 10 12 15 17 10"></polyline>
            <line x1="12" y1="15" x2="12" y2="3"></line>
          </svg>
          Install
        </template>
      </button>
    </div>

    <!-- Result messages -->
    <div v-if="result" class="result" :class="result.success ? 'result-success' : 'result-error'">
      <div class="result-content">
        <template v-if="result.success">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
          <div>
            <span class="result-title">
              Installed {{ result.skillsCount }} skill{{ result.skillsCount !== 1 ? 's' : '' }}
            </span>
            <span class="result-names">{{ result.skillNames.join(', ') }}</span>
          </div>
        </template>
        <template v-else>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="15" y1="9" x2="9" y2="15"></line>
            <line x1="9" y1="9" x2="15" y2="15"></line>
          </svg>
          <span>{{ result.errorMessage }}</span>
        </template>
      </div>
      <button class="dismiss-btn" @click="result = null">&times;</button>
    </div>

    <!-- Examples help -->
    <div v-if="!installing && !result" class="examples">
      <span class="examples-label">Examples</span>
      <div class="example-list">
        <button class="example" @click="fillExample('owner/repo')">
          <code>owner/repo</code>
          <span>GitHub shorthand</span>
        </button>
        <button class="example" @click="fillExample('owner/repo#branch')">
          <code>owner/repo#branch</code>
          <span>Specific branch or tag</span>
        </button>
        <button class="example" @click="fillExample('https://github.com/owner/repo')">
          <code>https://github.com/owner/repo</code>
          <span>Full GitHub URL</span>
        </button>
        <button class="example" @click="fillExample('https://example.com/skills.zip')">
          <code>https://example.com/skills.zip</code>
          <span>Zip archive URL</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AppService } from '../bindings/scribe'
import type { InstallResult } from '../types/skill'

const sourceStr = ref('')
const installing = ref(false)
const result = ref<InstallResult | null>(null)
const sourceInput = ref<HTMLInputElement | null>(null)

onMounted(() => {
  sourceInput.value?.focus()
})

async function handleInstall() {
  const source = sourceStr.value.trim()
  if (!source || installing.value) return

  installing.value = true
  result.value = null

  try {
    const res = await AppService.InstallFromSource(source)
    result.value = res as InstallResult
    if (res?.success) {
      sourceStr.value = ''
    }
  } catch (e) {
    result.value = {
      success: false,
      skillsCount: 0,
      skillNames: [],
      errorMessage: e instanceof Error ? e.message : 'Installation failed'
    }
  } finally {
    installing.value = false
  }
}

function fillExample(example: string) {
  sourceStr.value = example
  sourceInput.value?.focus()
}
</script>

<style scoped>
.install-skills {
  max-width: 560px;
}

.install-header {
  margin-bottom: 1.5rem;
}

.install-header h2 {
  font-size: 1rem;
  font-weight: 600;
  margin: 0 0 0.25rem 0;
}

.subtitle {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin: 0;
}

.install-form {
  display: flex;
  gap: 0.5rem;
}

.install-form input {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.875rem;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  font-family: inherit;
}

.install-form input:focus {
  outline: none;
  border-color: var(--accent-color);
}

.install-form input:disabled {
  opacity: 0.6;
}

.install-btn {
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

.install-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.install-btn:disabled {
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
  to {
    transform: rotate(360deg);
  }
}

/* Result messages */
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

.result-success {
  background-color: rgba(52, 199, 89, 0.1);
  color: var(--success-color);
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

.result-content div {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.result-title {
  font-weight: 600;
}

.result-names {
  font-size: 0.75rem;
  opacity: 0.8;
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

/* Examples */
.examples {
  margin-top: 1.5rem;
}

.examples-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.example-list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.example {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.375rem 0.5rem;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  text-align: left;
  transition: background-color 0.15s;
}

.example:hover {
  background-color: var(--bg-secondary);
}

.example code {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.75rem;
  color: var(--accent-color);
  min-width: 220px;
}

.example span {
  font-size: 0.75rem;
  color: var(--text-secondary);
}
</style>
