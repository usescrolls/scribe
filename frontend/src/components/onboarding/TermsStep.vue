<template>
  <div class="step terms-step">
    <h1>Terms & Conditions</h1>
    <p class="subtitle">Please review and accept before continuing</p>

    <div class="terms-box">
      <div class="terms-content">
        <template v-for="(clause, i) in clauses" :key="i">
          <h3>{{ i + 1 }}. {{ clause.title }}</h3>
          <p>{{ clause.body }}</p>
        </template>
      </div>
    </div>

    <label class="checkbox-label">
      <input
        v-model="accepted"
        type="checkbox"
        class="checkbox"
      />
      <span>I have read and agree to the Terms & Conditions</span>
    </label>

    <button
      class="primary-button"
      :disabled="!accepted"
      @click="$emit('accept')"
    >
      Continue
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AppService } from '../../bindings/scribe'

defineEmits<{
  accept: []
}>()

const accepted = ref(false)
const clauses = ref<{ title: string; body: string }[]>([])

onMounted(async () => {
  try {
    clauses.value = await AppService.GetTermsClauses()
  } catch {
    // Fallback: component will just show empty terms box
  }
})
</script>

<style scoped>
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 560px;
  width: 100%;
}

h1 {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.subtitle {
  font-size: 1rem;
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}

.terms-box {
  width: 100%;
  max-height: 320px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background-color: var(--bg-secondary);
  margin-bottom: 1.5rem;
  text-align: left;
}

.terms-content {
  padding: 1.25rem;
}

.terms-content h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.375rem 0;
}

.terms-content h3:not(:first-child) {
  margin-top: 1rem;
}

.terms-content p {
  font-size: 0.8125rem;
  line-height: 1.6;
  color: var(--text-secondary);
  margin: 0;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.checkbox {
  width: 1rem;
  height: 1rem;
  cursor: pointer;
  accent-color: var(--accent-color);
}

.primary-button {
  padding: 0.75rem 2rem;
  font-size: 1rem;
  font-weight: 500;
  color: white;
  background-color: var(--accent-color);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.primary-button:hover:not(:disabled) {
  opacity: 0.9;
}

.primary-button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
