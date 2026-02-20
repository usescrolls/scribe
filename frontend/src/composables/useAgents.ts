import { ref, computed, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import { useLogger } from "./useLogger"
import type { AgentStatus } from "../types/skill"

const log = useLogger("Agents")

export function useAgents() {
  const agents = ref<AgentStatus[]>([])
  const selectedAgent = ref<string | null>(null)
  const loading = ref(true)
  const error = ref<string | null>(null)
  let unsubscribeSkills: { (): void } | null = null
  let unsubscribeWorkspace: { (): void } | null = null

  async function fetchAgents() {
    try {
      loading.value = true
      error.value = null
      agents.value = await AppService.GetAgentStatus()
    } catch (e) {
      log.error(`failed to fetch agents: ${e instanceof Error ? e.message : e}`)
      error.value = e instanceof Error ? e.message : "Failed to load agents"
    } finally {
      loading.value = false
    }
  }

  function selectAgent(agentId: string | null) {
    selectedAgent.value = agentId
  }

  const installedAgents = computed(() =>
    agents.value.filter((a) => a.installed),
  )

  const installedCount = computed(() => installedAgents.value.length)
  const totalCount = computed(() => agents.value.length)

  onMounted(() => {
    fetchAgents()
    unsubscribeSkills = Events.On("skills-updated", fetchAgents)
    unsubscribeWorkspace = Events.On("workspace-changed", fetchAgents)
  })

  onUnmounted(() => {
    if (unsubscribeSkills) unsubscribeSkills()
    if (unsubscribeWorkspace) unsubscribeWorkspace()
  })

  return {
    agents,
    selectedAgent,
    loading,
    error,
    installedAgents,
    installedCount,
    totalCount,
    fetchAgents,
    selectAgent,
  }
}
