import { ref, computed, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import type { AgentStatus } from "../types/skill"

export function useAgents() {
  const agents = ref<AgentStatus[]>([])
  const selectedAgent = ref<string | null>(null)
  const loading = ref(true)
  const error = ref<string | null>(null)
  let unsubscribe: { (): void } | null = null

  async function fetchAgents() {
    try {
      loading.value = true
      error.value = null
      agents.value = await AppService.GetAgentStatus()
    } catch (e) {
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
    unsubscribe = Events.On("skills-updated", fetchAgents)
  })

  onUnmounted(() => {
    if (unsubscribe) {
      unsubscribe()
    }
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
