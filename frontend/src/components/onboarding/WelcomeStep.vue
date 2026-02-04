<template>
  <div class="step welcome-step">
    <img src="../../assets/icon.png" alt="Scribe" class="logo" />
    <h1>Welcome to Scribe</h1>
    <p class="tagline">
      Sync AI coding skills to all your coding agents.
      Install once, available everywhere.
    </p>
    <button class="primary-button" @click="$emit('next')">
      Get Started
    </button>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  next: []
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

.logo {
  width: 80px;
  height: 80px;
  margin-bottom: 1.5rem;
}

h1 {
  font-size: 2rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--text-primary);
}

.tagline {
  font-size: 1.125rem;
  color: var(--text-secondary);
  margin-bottom: 2rem;
  line-height: 1.6;
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
