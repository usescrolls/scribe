<template>
  <div class="install-skills">
    <div class="install-header">
      <h2>Install Skills</h2>
      <p class="subtitle">Add skills from GitHub, GitLab, Bitbucket, or a zip URL</p>
    </div>

    <!-- Step 1: Source Input -->
    <div v-if="step === 'source'" class="step-source">
      <div class="install-form">
        <input
          ref="sourceInput"
          v-model="sourceStr"
          type="text"
          placeholder="owner/repo, Git URL, zip URL, or paste a CLI command"
          autocorrect="off"
          autocomplete="off"
          autocapitalize="off"
          spellcheck="false"
          :disabled="discovering"
          @keyup.enter="handleDiscover"
        />
        <button
          class="install-btn"
          :disabled="!sourceStr.trim() || discovering"
          @click="handleDiscover"
        >
          <template v-if="discovering">
            <div class="spinner-sm"></div>
            {{ progressMessage || 'Fetching...' }}
          </template>
          <template v-else>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            Fetch
          </template>
        </button>
      </div>

      <!-- Error message -->
      <div v-if="error" class="result result-error">
        <div class="result-content">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="15" y1="9" x2="9" y2="15"></line>
            <line x1="9" y1="9" x2="15" y2="15"></line>
          </svg>
          <div>
            <span>{{ error.summary }}</span>
            <ul v-if="error.details" class="error-details">
              <li v-for="(detail, i) in error.details" :key="i">{{ detail }}</li>
            </ul>
            <span v-if="isAuthError(error.summary)" class="auth-hint">
              For private repos, ensure git credentials are configured (e.g. <code>gh auth login</code> for GitHub, or add your SSH key to ssh-agent).
            </span>
          </div>
        </div>
        <button class="dismiss-btn" @click="error = null">&times;</button>
      </div>

      <!-- Activity log during discovery -->
      <div v-if="discovering && activityLog.length > 0" class="activity-section">
        <button class="activity-toggle" @click="showLog = !showLog">
          {{ showLog ? 'Hide' : 'Show' }} activity
        </button>
        <div v-if="showLog" ref="activityLogEl" class="activity-log">
          <div
            v-for="(entry, i) in activityLog"
            :key="i"
            class="activity-entry"
            :class="{ done: entry.done }"
          >
            <span class="activity-icon">{{ entry.done ? '\u2713' : '\u2022' }}</span>
            <span>{{ entry.message }}</span>
          </div>
        </div>
      </div>

      <!-- Examples help -->
      <div v-if="!discovering && !error" class="examples">
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
          <button class="example" @click="fillExample('git@github.com:owner/repo.git')">
            <code>git@github.com:owner/repo.git</code>
            <span>SSH clone URL</span>
          </button>
          <button class="example" @click="fillExample('https://gitlab.com/group/subgroup/repo')">
            <code>https://gitlab.com/group/subgroup/repo</code>
            <span>GitLab nested groups</span>
          </button>
          <button class="example" @click="fillExample('https://bitbucket.org/owner/repo')">
            <code>https://bitbucket.org/owner/repo</code>
            <span>Bitbucket URL</span>
          </button>
          <button class="example" @click="fillExample('https://example.com/skills.zip')">
            <code>https://example.com/skills.zip</code>
            <span>Zip archive URL</span>
          </button>
          <button class="example" @click="fillExample('npx skills add owner/repo')">
            <code>npx skills add owner/repo</code>
            <span>Paste from CLI</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Step 2: Review Discovered Skills -->
    <div v-else-if="step === 'review'" class="step-review">
      <div class="step-info">
        <SourceAvatar :source="discoverResult!.source" :source-type="discoverResult!.sourceType" />
        <span class="source-badge">{{ discoverResult!.sourceType }}</span>
        <span class="source-label">{{ discoverResult!.source }}</span>
      </div>
      <p class="step-description">
        Found {{ discoverResult!.skills.length }} skill{{ discoverResult!.skills.length !== 1 ? 's' : '' }}.
        <template v-if="selectableSkills.length > 0">
          Select which ones to install:
        </template>
        <template v-else>
          All skills from this source are already installed.
        </template>
      </p>
      <div class="skill-checklist">
        <label v-if="selectableSkills.length > 0" class="select-all-item">
          <input
            type="checkbox"
            :checked="allSkillsSelected"
            @change="toggleAllSkills"
          />
          <span class="select-all-label">{{ allSkillsSelected ? 'Deselect all' : 'Select all' }}</span>
        </label>
        <label
          v-for="skill in discoverResult!.skills"
          :key="skill.name"
          class="skill-check-item"
          :class="{ 'skill-already-installed': skill.alreadyInstalled }"
        >
          <input
            type="checkbox"
            :checked="selectedSkills.has(skill.name)"
            :disabled="skill.alreadyInstalled"
            @change="toggleSkill(skill.name)"
          />
          <div class="skill-check-info">
            <span class="skill-check-name">{{ skill.name }}</span>
            <span v-if="skill.alreadyInstalled" class="already-installed-badge">already installed</span>
            <span v-else-if="skill.description" class="skill-check-desc">{{ skill.description }}</span>
          </div>
        </label>
      </div>
      <div class="step-actions">
        <button class="btn-secondary" @click="handleCancel">Cancel</button>
        <button
          class="btn-primary"
          :disabled="selectedSkills.size === 0"
          @click="step = 'workspaces'"
        >
          Continue
        </button>
      </div>
    </div>

    <!-- Step 3: Workspace Selection -->
    <div v-else-if="step === 'workspaces'" class="step-workspaces">
      <p class="step-description">
        Add to workspaces <span class="optional-hint">(optional)</span>
      </p>
      <div v-if="workspaces.length > 0" class="workspace-checklist">
        <label
          v-for="ws in workspaces"
          :key="ws.name"
          class="workspace-check-item"
        >
          <input
            type="checkbox"
            :checked="selectedWorkspaces.has(ws.name)"
            @change="toggleWorkspace(ws.name)"
          />
          <div class="workspace-check-info">
            <span class="workspace-check-name">
              {{ ws.name }}
              <span v-if="ws.isActive" class="active-badge">active</span>
            </span>
            <span class="workspace-check-desc">{{ ws.skills.length }} skill{{ ws.skills.length !== 1 ? 's' : '' }}</span>
          </div>
        </label>
      </div>
      <p v-else class="empty-hint">No workspaces found.</p>
      <div class="step-actions">
        <button class="btn-secondary" @click="step = 'review'">Back</button>
        <button
          class="btn-primary"
          @click="handleConfirmInstall"
        >
          Install {{ selectedSkills.size }} skill{{ selectedSkills.size !== 1 ? 's' : '' }}
        </button>
      </div>
    </div>

    <!-- Step 4: Installing Progress -->
    <div v-else-if="step === 'installing'" class="step-installing">
      <div class="installing-header">
        <div class="spinner"></div>
        <span>Installing skills...</span>
      </div>
      <div class="installing-progress">
        <div
          v-for="skillName in [...selectedSkills]"
          :key="skillName"
          class="installing-skill-item"
          :class="{
            done: installedSoFar.has(skillName),
            active: installingNow === skillName
          }"
        >
          <div class="installing-skill-icon">
            <svg v-if="installedSoFar.has(skillName)" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            <div v-else-if="installingNow === skillName" class="spinner-xs"></div>
            <div v-else class="pending-dot"></div>
          </div>
          <div class="installing-skill-info">
            <span class="installing-skill-name">{{ skillName }}</span>
            <span v-if="installingNow === skillName && progressDetail" class="installing-substep">{{ progressDetail }}</span>
          </div>
        </div>
      </div>
      <!-- Activity log during install -->
      <div v-if="activityLog.length > 0" class="activity-section">
        <button class="activity-toggle" @click="showLog = !showLog">
          {{ showLog ? 'Hide' : 'Show' }} activity
        </button>
        <div v-if="showLog" ref="activityLogEl" class="activity-log">
          <div
            v-for="(entry, i) in activityLog"
            :key="i"
            class="activity-entry"
            :class="{ done: entry.done }"
          >
            <span class="activity-icon">{{ entry.done ? '\u2713' : '\u2022' }}</span>
            <span>{{ entry.message }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Step 5: Result -->
    <div v-else-if="step === 'result'" class="step-result">
      <div v-if="installResult" class="result" :class="installResult.success ? 'result-success' : 'result-error'">
        <div class="result-content">
          <template v-if="installResult.success">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
              <polyline points="22 4 12 14.01 9 11.01"></polyline>
            </svg>
            <div>
              <span class="result-title">
                Installed {{ installResult.skillsCount }} skill{{ installResult.skillsCount !== 1 ? 's' : '' }}
              </span>
              <span class="result-names">{{ installResult.skillNames.join(', ') }}</span>
            </div>
          </template>
          <template v-else>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"></circle>
              <line x1="15" y1="9" x2="9" y2="15"></line>
              <line x1="9" y1="9" x2="15" y2="15"></line>
            </svg>
            <span>{{ installResult.errorMessage }}</span>
          </template>
        </div>
      </div>
      <button class="btn-secondary reset-btn" @click="resetToStart">Install more</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, onMounted, onUnmounted } from 'vue'
import SourceAvatar from './SourceAvatar.vue'
import { AppService } from '../bindings/scribe'
import { Events } from '@wailsio/runtime'
import type { DiscoverResult, InstallResult, WorkspaceInfo, ProgressEvent } from '../types/skill'

const props = defineProps<{
  initialSource?: string | null
}>()

const emit = defineEmits<{
  consumed: []
}>()

type Step = 'source' | 'review' | 'workspaces' | 'installing' | 'result'

const sourceStr = ref('')
const step = ref<Step>('source')
const discovering = ref(false)
interface DiscoverError {
  summary: string
  details?: string[]
}
const error = ref<DiscoverError | null>(null)
const sourceInput = ref<HTMLInputElement | null>(null)

const discoverResult = ref<DiscoverResult | null>(null)
const selectedSkills = reactive(new Set<string>())
const workspaces = ref<WorkspaceInfo[]>([])
const selectedWorkspaces = reactive(new Set<string>())
const installResult = ref<InstallResult | null>(null)

// Per-skill install progress
const installingNow = ref<string | null>(null)
const installedSoFar = reactive(new Set<string>())

// Progress feedback
const progressMessage = ref<string | null>(null)
const progressDetail = ref<string | null>(null)
const activityLog = reactive<{ message: string; done?: boolean }[]>([])
const showLog = ref(false)
const activityLogEl = ref<HTMLElement | null>(null)

let unsubscribeWorkspace: { (): void } | null = null
let unsubscribeProgress: { (): void } | null = null
let unsubscribeProgressEvent: { (): void } | null = null

onMounted(() => {
  sourceInput.value?.focus()
  unsubscribeWorkspace = Events.On('workspace-changed', fetchWorkspaces)
  unsubscribeProgress = Events.On('install-progress', (event: { data: { skillName: string; current: number; total: number } }) => {
    const { skillName } = event.data
    // Mark the previous skill as done
    if (installingNow.value && installingNow.value !== skillName) {
      installedSoFar.add(installingNow.value)
    }
    installingNow.value = skillName
  })
  unsubscribeProgressEvent = Events.On('progress', (event: { data: ProgressEvent }) => {
    const e = event.data
    progressMessage.value = e.message

    // Mark previous log entry as done when a new step arrives
    if (activityLog.length > 0) {
      activityLog[activityLog.length - 1].done = true
    }

    // Update sub-step detail for install phase
    if (e.phase === 'install') {
      if (e.step === 'copying' || e.step === 'metadata' || e.step === 'syncing' || e.step === 'workspace') {
        progressDetail.value = e.message
      } else if (e.step === 'done' || e.step === 'error') {
        progressDetail.value = null
      }
    }

    // Add to activity log
    activityLog.push({ message: e.message, done: e.step === 'done' })

    // Auto-scroll log to bottom
    nextTick(() => {
      if (activityLogEl.value) {
        activityLogEl.value.scrollTop = activityLogEl.value.scrollHeight
      }
    })
  })
  fetchWorkspaces()

  // If mounted with an initial source (e.g. from Marketplace), auto-trigger
  if (props.initialSource) {
    sourceStr.value = props.initialSource
    emit('consumed')
    handleDiscover()
  }
})

onUnmounted(() => {
  if (unsubscribeWorkspace) unsubscribeWorkspace()
  if (unsubscribeProgress) unsubscribeProgress()
  if (unsubscribeProgressEvent) unsubscribeProgressEvent()
})

async function fetchWorkspaces() {
  try {
    workspaces.value = await AppService.GetWorkspaces()
  } catch {
    // Non-critical, workspaces just won't show
  }
}

async function handleDiscover() {
  const source = parseSourceInput(sourceStr.value.trim())
  if (!source || discovering.value) return
  sourceStr.value = source

  discovering.value = true
  error.value = null
  progressMessage.value = null
  progressDetail.value = null
  activityLog.length = 0

  try {
    const res = await AppService.DiscoverFromSource(source)
    discoverResult.value = res as DiscoverResult

    // Select only new (not already installed) skills by default
    selectedSkills.clear()
    for (const skill of discoverResult.value.skills) {
      if (!skill.alreadyInstalled) {
        selectedSkills.add(skill.name)
      }
    }

    // No workspaces pre-selected — user opts in via checkboxes
    selectedWorkspaces.clear()

    // Refresh workspaces for step 3
    await fetchWorkspaces()

    step.value = 'review'
  } catch (e) {
    error.value = parseDiscoverError(e)
  } finally {
    discovering.value = false
  }
}

const selectableSkills = computed(() => {
  if (!discoverResult.value) return []
  return discoverResult.value.skills.filter(s => !s.alreadyInstalled)
})

const allSkillsSelected = computed(() => {
  if (selectableSkills.value.length === 0) return false
  return selectableSkills.value.every(s => selectedSkills.has(s.name))
})

function toggleAllSkills() {
  if (allSkillsSelected.value) {
    for (const skill of selectableSkills.value) {
      selectedSkills.delete(skill.name)
    }
  } else {
    for (const skill of selectableSkills.value) {
      selectedSkills.add(skill.name)
    }
  }
}

function toggleSkill(name: string) {
  if (selectedSkills.has(name)) {
    selectedSkills.delete(name)
  } else {
    selectedSkills.add(name)
  }
}

function toggleWorkspace(name: string) {
  if (selectedWorkspaces.has(name)) {
    selectedWorkspaces.delete(name)
  } else {
    selectedWorkspaces.add(name)
  }
}

async function handleConfirmInstall() {
  if (step.value === 'installing') return

  // Reset progress state and transition to installing step
  installingNow.value = null
  installedSoFar.clear()
  progressMessage.value = null
  progressDetail.value = null
  activityLog.length = 0
  error.value = null
  step.value = 'installing'

  try {
    const res = await AppService.ConfirmInstall(
      [...selectedSkills],
      [...selectedWorkspaces]
    )
    // Mark the last skill as done
    if (installingNow.value) {
      installedSoFar.add(installingNow.value)
    }
    installResult.value = res as InstallResult
    step.value = 'result'
  } catch (e) {
    error.value = { summary: extractErrorMessage(e) || 'Installation failed' }
    step.value = 'source'
  }
}

async function handleCancel() {
  try {
    await AppService.CancelDiscover()
  } catch {
    // Ignore cleanup errors
  }
  resetToStart()
}

function resetToStart() {
  step.value = 'source'
  sourceStr.value = ''
  discoverResult.value = null
  selectedSkills.clear()
  selectedWorkspaces.clear()
  installResult.value = null
  error.value = null
  progressMessage.value = null
  progressDetail.value = null
  activityLog.length = 0
  showLog.value = false
  setTimeout(() => sourceInput.value?.focus(), 50)
}

function fillExample(example: string) {
  sourceStr.value = example
  sourceInput.value?.focus()
}

function parseSourceInput(input: string): string {
  const trimmed = input.trim()
  const skillsMatch = trimmed.match(/^(?:npx\s+)?skills\s+add\s+(.+)$/i)
  if (skillsMatch) return firstCommandArg(skillsMatch[1]) || trimmed

  const scribeMatch = trimmed.match(/^scribe\s+install\s+(.+)$/i)
  if (scribeMatch) return firstCommandArg(scribeMatch[1]) || trimmed

  const gitCloneMatch = trimmed.match(/^git\s+clone\s+(.+)$/i)
  if (gitCloneMatch) return firstCommandArg(gitCloneMatch[1]) || trimmed

  const ghCloneMatch = trimmed.match(/^gh\s+repo\s+clone\s+(.+)$/i)
  if (ghCloneMatch) return firstCommandArg(ghCloneMatch[1]) || trimmed

  return trimmed
}

function firstCommandArg(rest: string): string {
  const parts = rest.match(/"[^"]+"|'[^']+'|\S+/g) ?? []
  let skipNext = false
  for (const rawPart of parts) {
    if (skipNext) {
      skipNext = false
      continue
    }
    const part = stripQuotes(rawPart)
    if (part.startsWith('-')) {
      skipNext = commandFlagTakesValue(part)
      continue
    }
    return part
  }
  return ''
}

function commandFlagTakesValue(flag: string): boolean {
  if (flag.includes('=')) return false
  return [
    '-b', '--branch', '-c', '--config', '--depth', '-o', '--origin',
    '--reference', '--reference-if-able', '--separate-git-dir',
    '--template', '-u', '--upload-pack',
  ].includes(flag)
}

function stripQuotes(value: string): string {
  if (value.length >= 2) {
    const first = value[0]
    const last = value[value.length - 1]
    if ((first === '"' && last === '"') || (first === "'" && last === "'")) {
      return value.slice(1, -1)
    }
  }
  return value
}

function extractErrorMessage(e: unknown): string {
  // Get the raw message string
  let msg: string
  if (e instanceof Error) msg = e.message
  else if (typeof e === 'object' && e !== null && 'message' in e) msg = String((e as { message: unknown }).message)
  else msg = String(e)

  // Wails wraps Go errors as new Error(jsonString) — unwrap the JSON envelope
  if (msg.startsWith('{')) {
    try {
      const parsed = JSON.parse(msg)
      if (parsed && typeof parsed.message === 'string') return parsed.message
    } catch { /* not JSON, use as-is */ }
  }
  return msg
}

function parseDiscoverError(e: unknown): DiscoverError {
  let msg = extractErrorMessage(e)
  // Strip redundant prefix
  msg = msg.replace(/^failed to fetch skills:\s*/i, '')

  // Parse structured "found N SKILL.md file(s) but none could be parsed:" errors
  const match = msg.match(/^(found \d+ SKILL\.md .+?parsed):?\s*(.*)$/s)
  if (match) {
    const details = match[2]
      .split('\n')
      .map(line => line.trim())
      .filter(Boolean)
      .map(line => {
        // "path/SKILL.md: error message" -> "path/SKILL.md — error message"
        return line.replace(/:\s*/, ' \u2014 ')
      })
    return { summary: match[1], details }
  }

  return { summary: msg }
}

function isAuthError(msg: string): boolean {
  const lower = msg.toLowerCase()
  return ['authentication required', 'authentication failed', 'permission denied',
    'repository not found', 'access denied', '403', '401'].some(p => lower.includes(p))
}
</script>

<style scoped>
.install-skills {
  max-width: 560px;
  margin: 0 auto;
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

/* Source input form */
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

.error-details {
  margin: 0.375rem 0 0 0;
  padding: 0;
  list-style: none;
  font-size: 0.75rem;
  opacity: 0.85;
}

.error-details li {
  padding: 0.125rem 0;
}

.result-title {
  font-weight: 600;
}

.result-names {
  font-size: 0.75rem;
  opacity: 0.8;
}

.auth-hint {
  display: block;
  margin-top: 0.375rem;
  font-size: 0.75rem;
  opacity: 0.85;
}

.auth-hint code {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.6875rem;
  padding: 0.0625rem 0.25rem;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
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
  min-width: 270px;
}

.example span {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

/* Step info bar */
.step-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.source-badge {
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 3px;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
}

.source-label {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.step-description {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin: 0 0 1rem 0;
}

.optional-hint {
  opacity: 0.6;
}

/* Skill checklist */
.skill-checklist {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 1rem;
}

.select-all-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.375rem 0.75rem;
  cursor: pointer;
  margin-bottom: 0.125rem;
}

.select-all-item input[type="checkbox"] {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
}

.select-all-label {
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.skill-check-item {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.15s;
}

.skill-check-item:hover {
  border-color: var(--accent-color);
}

.skill-check-item input[type="checkbox"] {
  margin-top: 0.125rem;
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
}

.skill-check-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.skill-check-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.skill-check-desc {
  font-size: 0.75rem;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-check-item.skill-already-installed {
  opacity: 0.55;
  cursor: default;
}

.skill-check-item.skill-already-installed:hover {
  border-color: var(--border-color);
}

.skill-check-item.skill-already-installed input[type="checkbox"] {
  cursor: default;
}

.already-installed-badge {
  font-size: 0.6875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

/* Workspace checklist */
.workspace-checklist {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  margin-bottom: 1rem;
}

.workspace-check-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.15s;
}

.workspace-check-item:hover {
  border-color: var(--accent-color);
}

.workspace-check-item input[type="checkbox"] {
  width: 0.875rem;
  height: 0.875rem;
  cursor: pointer;
  flex-shrink: 0;
}

.workspace-check-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.workspace-check-name {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.active-badge {
  font-size: 0.5625rem;
  font-weight: 600;
  text-transform: uppercase;
  padding: 0.0625rem 0.25rem;
  background-color: rgba(52, 199, 89, 0.15);
  color: var(--success-color);
  border-radius: 3px;
}

.workspace-check-desc {
  font-size: 0.6875rem;
  color: var(--text-secondary);
}

.empty-hint {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin-bottom: 1rem;
}

/* Step action buttons */
.step-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}

.btn-primary {
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
  transition: opacity 0.15s;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 500;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.15s;
}

.btn-secondary:hover {
  background-color: var(--bg-primary);
}

.reset-btn {
  margin-top: 1rem;
}

/* Installing progress step */
.installing-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin-bottom: 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.installing-progress {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.installing-skill-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.625rem;
  border-radius: 6px;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.installing-skill-item.active {
  color: var(--text-primary);
  background-color: var(--bg-secondary);
}

.installing-skill-item.done {
  color: var(--success-color);
}

.installing-skill-icon {
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.installing-skill-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}

.installing-skill-name {
  font-weight: 500;
}

.installing-substep {
  font-size: 0.6875rem;
  color: var(--text-secondary);
  opacity: 0.8;
}

/* Activity log */
.activity-section {
  margin-top: 0.75rem;
}

.activity-toggle {
  padding: 0;
  border: none;
  background: transparent;
  font-size: 0.6875rem;
  color: var(--text-secondary);
  cursor: pointer;
  text-decoration: underline;
  text-decoration-style: dotted;
  text-underline-offset: 2px;
  opacity: 0.8;
  transition: opacity 0.15s;
}

.activity-toggle:hover {
  opacity: 1;
}

.activity-log {
  margin-top: 0.375rem;
  max-height: 160px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 0.375rem 0.5rem;
  background-color: var(--bg-secondary);
  font-size: 0.6875rem;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.activity-entry {
  display: flex;
  align-items: baseline;
  gap: 0.375rem;
  padding: 0.0625rem 0;
  color: var(--text-secondary);
}

.activity-entry.done {
  color: var(--success-color);
}

.activity-icon {
  flex-shrink: 0;
  width: 0.75rem;
  text-align: center;
}

.spinner-xs {
  width: 12px;
  height: 12px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.pending-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background-color: var(--border-color);
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
</style>
