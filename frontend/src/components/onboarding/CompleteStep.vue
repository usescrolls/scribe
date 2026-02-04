<template>
  <div class="step complete-step">
    <div class="success-icon">✓</div>
    <h1>You're All Set!</h1>
    <p class="description">
      Scribe is ready to manage your AI coding skills.
      Your demo skill has been synced to all your coding agents.
    </p>

    <div class="next-steps">
      <h3>Next Steps</h3>
      <ul>
        <li>Visit AgentHub to discover more skills</li>
        <li>Use <code>scribe install &lt;repo&gt;</code> to install skills from GitHub</li>
        <li>Create your own skills with SKILL.md files</li>
      </ul>
    </div>

    <button class="primary-button" @click="$emit('finish')">
      Open Scribe
    </button>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  finish: []
}>()
</script>

<style scoped>
.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 480px;
  width: 100%;
}

.success-icon {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  background-color: #22c55e;
  color: white;
  border-radius: 50%;
  margin-bottom: 1.5rem;
}

h1 {
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--text-primary);
}

.description {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.next-steps {
  width: 100%;
  padding: 1rem;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  text-align: left;
  margin-bottom: 1.5rem;
}

.next-steps h3 {
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
  color: var(--text-primary);
}

.next-steps ul {
  margin: 0;
  padding-left: 1.25rem;
}

.next-steps li {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
  line-height: 1.5;
}

.next-steps li:last-child {
  margin-bottom: 0;
}

.next-steps code {
  font-size: 0.8125rem;
  padding: 0.125rem 0.375rem;
  background-color: var(--bg-primary);
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
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

.primary-button:hover {
  opacity: 0.9;
}
</style>
