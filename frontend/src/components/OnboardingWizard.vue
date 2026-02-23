<template>
  <div class="onboarding-wizard">
    <div class="wizard-content">
      <Transition name="fade" mode="out-in">
        <WelcomeStep
          v-if="currentStep === 'welcome'"
          @next="goToTerms"
        />
        <TermsStep
          v-else-if="currentStep === 'terms'"
          @accept="handleAcceptTerms"
        />
        <AgentDetectionStep
          v-else-if="currentStep === 'agents'"
          :agents="agents"
          :loading="agentsLoading"
          :has-installed-agents="hasInstalledAgents"
          @next="goToExistingSkills"
        />
        <ExistingSkillsStep
          v-else-if="currentStep === 'existing-skills'"
          :skills="existingSkills"
          :conflicts="skillConflicts"
          :loading="existingSkillsLoading"
          @import-all="handleImportAll"
          @delete-all="handleDeleteAll"
          @resolve-conflict="handleResolveConflict"
          @next="goToInstallDemo"
        />
        <InstallDemoStep
          v-else-if="currentStep === 'install-demo'"
          :installing="demoSkillInstalling"
          :installed-agents="installedAgents"
          @install="handleInstallDemo"
          @skip="handleSkipDemo"
        />
        <CompleteStep
          v-else-if="currentStep === 'complete'"
          @finish="handleFinish"
        />
      </Transition>
    </div>

    <div class="wizard-footer">
      <div class="progress-dots">
        <span
          v-for="step in steps"
          :key="step"
          class="dot"
          :class="{ active: step === currentStep, completed: isStepCompleted(step) }"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useOnboarding, type OnboardingStep } from '../composables/useOnboarding'
import WelcomeStep from './onboarding/WelcomeStep.vue'
import TermsStep from './onboarding/TermsStep.vue'
import AgentDetectionStep from './onboarding/AgentDetectionStep.vue'
import ExistingSkillsStep from './onboarding/ExistingSkillsStep.vue'
import InstallDemoStep from './onboarding/InstallDemoStep.vue'
import CompleteStep from './onboarding/CompleteStep.vue'

const emit = defineEmits<{
  complete: []
}>()

const {
  currentStep,
  agents,
  agentsLoading,
  existingSkills,
  skillConflicts,
  existingSkillsLoading,
  demoSkillInstalling,
  installedAgents,
  hasInstalledAgents,
  goToStep,
  startAgentScan,
  stopAgentScan,
  fetchExistingSkills,
  importAllSkills,
  deleteAllSkills,
  resolveConflict,
  installDemoSkill,
  acceptTerms,
  completeOnboarding,
} = useOnboarding()

const steps: OnboardingStep[] = ['welcome', 'terms', 'agents', 'existing-skills', 'install-demo', 'complete']

const currentStepIndex = computed(() => steps.indexOf(currentStep.value))

function isStepCompleted(step: OnboardingStep): boolean {
  return steps.indexOf(step) < currentStepIndex.value
}

function goToTerms() {
  goToStep('terms')
}

async function handleAcceptTerms() {
  await acceptTerms()
  goToStep('agents')
  startAgentScan()
}

async function goToExistingSkills() {
  stopAgentScan()
  goToStep('existing-skills')
  await fetchExistingSkills()
  if (existingSkills.value.length === 0) {
    goToInstallDemo()
  }
}

function goToInstallDemo() {
  goToStep('install-demo')
}

async function handleImportAll() {
  const success = await importAllSkills()
  if (success) {
    goToInstallDemo()
  }
}

async function handleDeleteAll() {
  const success = await deleteAllSkills()
  if (success) {
    goToInstallDemo()
  }
}

async function handleResolveConflict(skillPath: string) {
  await resolveConflict(skillPath)
}

async function handleInstallDemo() {
  const success = await installDemoSkill()
  if (success) {
    goToStep('complete')
  }
}

function handleSkipDemo() {
  goToStep('complete')
}

async function handleFinish() {
  await completeOnboarding()
  emit('complete')
}

// Skip existing skills step if no skills found
watch(existingSkills, (skills) => {
  if (currentStep.value === 'existing-skills' && skills.length === 0 && !existingSkillsLoading.value) {
    goToInstallDemo()
  }
})
</script>

<style scoped>
.onboarding-wizard {
  position: fixed;
  inset: 0;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-primary);
  z-index: 1000;
}

.wizard-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  overflow-y: auto;
}

.wizard-footer {
  padding: 1.5rem;
  display: flex;
  justify-content: center;
  border-top: 1px solid var(--border-color);
  background-color: var(--bg-secondary);
}

.progress-dots {
  display: flex;
  gap: 0.5rem;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--border-color);
  transition: background-color 0.2s ease;
}

.dot.active {
  background-color: var(--accent-color);
}

.dot.completed {
  background-color: var(--text-secondary);
}

/* Fade transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
