import { ref, computed, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import { useLogger } from "./useLogger"
import type {
  AgentStatus,
  ExistingSkillInfo,
  SkillConflict,
} from "../types/skill"

const log = useLogger("Onboarding")

export type OnboardingStep = "welcome" | "agents" | "existing-skills" | "install-demo" | "complete"

export function useOnboarding() {
  const isCompleted = ref(false)
  const currentStep = ref<OnboardingStep>("welcome")
  const loading = ref(true)
  const error = ref<string | null>(null)

  // Agent detection state
  const agents = ref<AgentStatus[]>([])
  const agentsLoading = ref(false)

  // Existing skills state
  const existingSkills = ref<ExistingSkillInfo[]>([])
  const skillConflicts = ref<SkillConflict[]>([])
  const existingSkillsLoading = ref(false)

  // Demo skill installation state
  const demoSkillInstalling = ref(false)

  let unsubscribeCompleted: { (): void } | null = null
  let agentScanInterval: ReturnType<typeof setInterval> | null = null

  // Computed properties
  const installedAgents = computed(() =>
    agents.value.filter((a) => a.installed),
  )
  const hasInstalledAgents = computed(() => installedAgents.value.length > 0)
  const hasExistingSkills = computed(() => existingSkills.value.length > 0)
  const hasConflicts = computed(() => skillConflicts.value.length > 0)

  // Check if onboarding is completed
  async function checkOnboardingStatus() {
    try {
      loading.value = true
      isCompleted.value = await AppService.IsOnboardingCompleted()
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to check onboarding status"
    } finally {
      loading.value = false
    }
  }

  // Fetch agents (used for auto-scan every 30s)
  async function fetchAgents() {
    try {
      agentsLoading.value = true
      agents.value = await AppService.GetAgentStatus()
    } catch (e) {
      log.error(`failed to fetch agents: ${e instanceof Error ? e.message : e}`)
    } finally {
      agentsLoading.value = false
    }
  }

  // Start agent scanning (every 30 seconds)
  function startAgentScan() {
    fetchAgents()
    agentScanInterval = setInterval(fetchAgents, 30000)
  }

  // Stop agent scanning
  function stopAgentScan() {
    if (agentScanInterval) {
      clearInterval(agentScanInterval)
      agentScanInterval = null
    }
  }

  // Fetch existing skills in agent directories
  async function fetchExistingSkills() {
    try {
      existingSkillsLoading.value = true
      existingSkills.value = await AppService.DetectExistingSkills()
      skillConflicts.value = await AppService.DetectSkillConflicts()
    } catch (e) {
      log.error(
        `failed to detect existing skills: ${
          e instanceof Error ? e.message : e
        }`,
      )
    } finally {
      existingSkillsLoading.value = false
    }
  }

  // Import all existing skills
  async function importAllSkills(): Promise<boolean> {
    log.info("importing all existing skills")
    try {
      existingSkillsLoading.value = true
      await AppService.ImportAllExistingSkills()
      existingSkills.value = []
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to import skills"
      return false
    } finally {
      existingSkillsLoading.value = false
    }
  }

  // Delete all existing skills
  async function deleteAllSkills(): Promise<boolean> {
    try {
      existingSkillsLoading.value = true
      await AppService.DeleteAllExistingSkills()
      existingSkills.value = []
      return true
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to delete skills"
      return false
    } finally {
      existingSkillsLoading.value = false
    }
  }

  // Resolve a skill conflict by choosing a specific version
  async function resolveConflict(skillPath: string): Promise<boolean> {
    try {
      await AppService.ResolveSkillConflict(skillPath)
      // Refresh existing skills list
      await fetchExistingSkills()
      return true
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to resolve conflict"
      return false
    }
  }

  // Install demo skill
  async function installDemoSkill(): Promise<boolean> {
    try {
      demoSkillInstalling.value = true
      await AppService.InstallDemoSkill()
      return true
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to install demo skill"
      return false
    } finally {
      demoSkillInstalling.value = false
    }
  }

  // Complete onboarding
  async function completeOnboarding(): Promise<boolean> {
    log.info("completing onboarding")
    try {
      await AppService.CompleteOnboarding()
      isCompleted.value = true
      return true
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to complete onboarding"
      return false
    }
  }

  // Navigation helpers
  function goToStep(step: OnboardingStep) {
    currentStep.value = step
  }

  function nextStep() {
    const steps: OnboardingStep[] = [
      "welcome",
      "agents",
      "existing-skills",
      "install-demo",
      "complete",
    ]
    const currentIndex = steps.indexOf(currentStep.value)
    if (currentIndex < steps.length - 1) {
      currentStep.value = steps[currentIndex + 1]
    }
  }

  function previousStep() {
    const steps: OnboardingStep[] = [
      "welcome",
      "agents",
      "existing-skills",
      "install-demo",
      "complete",
    ]
    const currentIndex = steps.indexOf(currentStep.value)
    if (currentIndex > 0) {
      currentStep.value = steps[currentIndex - 1]
    }
  }

  onMounted(() => {
    checkOnboardingStatus()
    unsubscribeCompleted = Events.On("onboarding-completed", () => {
      isCompleted.value = true
    })
  })

  onUnmounted(() => {
    if (unsubscribeCompleted) {
      unsubscribeCompleted()
    }
    stopAgentScan()
  })

  return {
    // State
    isCompleted,
    currentStep,
    loading,
    error,
    agents,
    agentsLoading,
    existingSkills,
    skillConflicts,
    existingSkillsLoading,
    demoSkillInstalling,

    // Computed
    installedAgents,
    hasInstalledAgents,
    hasExistingSkills,
    hasConflicts,

    // Actions
    checkOnboardingStatus,
    fetchAgents,
    startAgentScan,
    stopAgentScan,
    fetchExistingSkills,
    importAllSkills,
    deleteAllSkills,
    resolveConflict,
    installDemoSkill,
    completeOnboarding,

    // Navigation
    goToStep,
    nextStep,
    previousStep,
  }
}
