<template>
  <Teleport to="body">
    <Transition name="modal" appear @after-leave="onAfterLeave">
      <div v-if="visible" class="detail-backdrop" @click.self="handleClose">
        <div class="detail-modal">
          <div class="detail-header">
            <div class="detail-title">
              <span class="source-badge">{{ skill.sourceType }}</span>
              <h3>{{ displayName }}</h3>
            </div>
            <button class="close-btn" @click="handleClose" title="Close">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          </div>
          <p v-if="skill.description" class="detail-description">{{ skill.description }}</p>
          <div class="detail-body">
            <div v-if="loading" class="loading">
              <div class="spinner"></div>
              <span>Loading content...</span>
            </div>
            <div v-else-if="error" class="error">
              <span>{{ error }}</span>
            </div>
            <div v-else-if="!renderedContent" class="empty">
              <span>No content available for this skill.</span>
            </div>
            <div v-else ref="contentEl" class="markdown-content" v-html="renderedContent"></div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted } from 'vue'
import { Browser } from '@wailsio/runtime'
import { marked } from 'marked'
import { AppService } from '../bindings/scribe'
import type { SkillInfo } from '../types/skill'

const props = defineProps<{
  skill: SkillInfo
}>()

const emit = defineEmits<{
  close: []
}>()

const visible = ref(true)
const loading = ref(true)
const error = ref<string | null>(null)
const renderedContent = ref<string | null>(null)
const contentEl = ref<HTMLElement | null>(null)
const displayName = computed(() => props.skill.displayName || props.skill.name)

// Intercept link clicks in rendered markdown and open in system browser
function handleContentClick(e: MouseEvent) {
  const target = (e.target as HTMLElement).closest('a')
  if (!target) return
  const href = target.getAttribute('href')
  if (!href) return
  e.preventDefault()
  if (href.startsWith('http://') || href.startsWith('https://')) {
    Browser.OpenURL(href)
  }
}

watch(contentEl, (el) => {
  if (el) el.addEventListener('click', handleContentClick)
}, { flush: 'post' })

async function loadContent() {
  try {
    loading.value = true
    error.value = null
    const content = await AppService.GetSkillContent(props.skill.name)
    if (content) {
      renderedContent.value = await marked(content)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load skill content'
  } finally {
    loading.value = false
  }
}

function handleClose() {
  visible.value = false
}

function onAfterLeave() {
  emit('close')
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    handleClose()
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  loadContent()
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.detail-backdrop {
  position: fixed;
  inset: 0;
  z-index: 200;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-modal {
  width: 600px;
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

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem 0;
  flex-shrink: 0;
}

.detail-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.detail-title h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.source-badge {
  flex-shrink: 0;
  padding: 0.125rem 0.375rem;
  background-color: var(--accent-color);
  color: white;
  border-radius: 3px;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.close-btn {
  width: 1.5rem;
  height: 1.5rem;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex-shrink: 0;
}

.close-btn:hover {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.detail-description {
  margin: 0;
  padding: 0.5rem 1.25rem 0;
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.5;
  flex-shrink: 0;
}

.detail-body {
  padding: 1rem 1.25rem 1.25rem;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
}

.loading,
.error,
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 2rem;
  color: var(--text-secondary);
  font-size: 0.8125rem;
}

.error {
  color: var(--danger-color);
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Markdown content styling */
.markdown-content {
  font-size: 0.8125rem;
  line-height: 1.6;
  color: var(--text-primary);
  word-wrap: break-word;
}

.markdown-content :deep(h1) {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 0.75rem;
  padding-bottom: 0.375rem;
  border-bottom: 1px solid var(--border-color);
}

.markdown-content :deep(h2) {
  font-size: 1.0625rem;
  font-weight: 600;
  margin: 1.25rem 0 0.5rem;
}

.markdown-content :deep(h3) {
  font-size: 0.9375rem;
  font-weight: 600;
  margin: 1rem 0 0.375rem;
}

.markdown-content :deep(p) {
  margin: 0 0 0.75rem;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 0 0 0.75rem;
  padding-left: 1.5rem;
}

.markdown-content :deep(li) {
  margin-bottom: 0.25rem;
}

.markdown-content :deep(code) {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.75rem;
  background-color: var(--bg-secondary);
  padding: 0.125rem 0.3125rem;
  border-radius: 3px;
  border: 1px solid var(--border-color);
}

.markdown-content :deep(pre) {
  margin: 0 0 0.75rem;
  padding: 0.75rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  overflow-x: auto;
}

.markdown-content :deep(pre code) {
  background: none;
  padding: 0;
  border: none;
  font-size: 0.75rem;
}

.markdown-content :deep(blockquote) {
  margin: 0 0 0.75rem;
  padding: 0.5rem 0.75rem;
  border-left: 3px solid var(--accent-color);
  background-color: var(--bg-secondary);
  border-radius: 0 4px 4px 0;
}

.markdown-content :deep(blockquote p) {
  margin: 0;
}

.markdown-content :deep(a) {
  color: var(--accent-color);
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

.markdown-content :deep(hr) {
  border: none;
  border-top: 1px solid var(--border-color);
  margin: 1rem 0;
}

.markdown-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 0 0 0.75rem;
  font-size: 0.75rem;
}

.markdown-content :deep(th),
.markdown-content :deep(td) {
  padding: 0.375rem 0.625rem;
  border: 1px solid var(--border-color);
  text-align: left;
}

.markdown-content :deep(th) {
  background-color: var(--bg-secondary);
  font-weight: 600;
}

.markdown-content :deep(strong) {
  font-weight: 600;
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .detail-modal,
.modal-leave-active .detail-modal {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .detail-modal {
  transform: scale(0.95);
  opacity: 0;
}

.modal-leave-to .detail-modal {
  transform: scale(0.95);
  opacity: 0;
}
</style>
