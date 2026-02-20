import { ref, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import { useLogger } from "./useLogger"
import type { WorkspaceInfo } from "../types/skill"

const log = useLogger("Workspaces")

const SKIP_SWITCH_INFO_KEY = "scribe-skip-workspace-switch-info"

// Shared across all composable instances
export const showSwitchInfoModal = ref(false)
export const switchInfoWorkspaceName = ref("")

export function dismissSwitchInfo(dontShowAgain: boolean) {
  showSwitchInfoModal.value = false
  if (dontShowAgain) {
    try {
      localStorage.setItem(SKIP_SWITCH_INFO_KEY, "true")
    } catch {
      // localStorage may be unavailable
    }
  }
}

function shouldShowSwitchInfo(): boolean {
  try {
    return localStorage.getItem(SKIP_SWITCH_INFO_KEY) !== "true"
  } catch {
    return false
  }
}

export function useWorkspaces() {
  const workspaces = ref<WorkspaceInfo[]>([])
  const activeWorkspace = ref<string>("default")
  const loading = ref(true)
  const error = ref<string | null>(null)
  let unsubscribe: { (): void } | null = null
  let unsubscribeSkills: { (): void } | null = null

  async function fetchWorkspaces() {
    try {
      loading.value = true
      error.value = null
      workspaces.value = await AppService.GetWorkspaces()
      const active = workspaces.value.find((ws) => ws.isActive)
      if (active) {
        activeWorkspace.value = active.name
      }
    } catch (e) {
      log.error(
        `failed to fetch workspaces: ${e instanceof Error ? e.message : e}`,
      )
      error.value = e instanceof Error ? e.message : "Failed to load workspaces"
    } finally {
      loading.value = false
    }
  }

  async function switchWorkspace(name: string): Promise<boolean> {
    log.info(`switching workspace to: ${name}`)
    try {
      await AppService.SetActiveWorkspace(name)
      activeWorkspace.value = name
      await fetchWorkspaces()
      if (shouldShowSwitchInfo()) {
        switchInfoWorkspaceName.value = name
        showSwitchInfoModal.value = true
      }
      return true
    } catch (e) {
      log.error(
        `failed to switch workspace to ${name}: ${
          e instanceof Error ? e.message : e
        }`,
      )
      error.value =
        e instanceof Error ? e.message : "Failed to switch workspace"
      return false
    }
  }

  async function createWorkspace(
    name: string,
    description: string = "",
  ): Promise<boolean> {
    log.info(`creating workspace: ${name}`)
    try {
      await AppService.CreateWorkspace(name, description)
      await fetchWorkspaces()
      return true
    } catch (e) {
      log.error(
        `failed to create workspace ${name}: ${
          e instanceof Error ? e.message : e
        }`,
      )
      error.value =
        e instanceof Error ? e.message : "Failed to create workspace"
      return false
    }
  }

  async function deleteWorkspace(name: string): Promise<boolean> {
    log.info(`deleting workspace: ${name}`)
    try {
      await AppService.DeleteWorkspace(name)
      await fetchWorkspaces()
      return true
    } catch (e) {
      log.error(
        `failed to delete workspace ${name}: ${
          e instanceof Error ? e.message : e
        }`,
      )
      error.value =
        e instanceof Error ? e.message : "Failed to delete workspace"
      return false
    }
  }

  onMounted(() => {
    fetchWorkspaces()
    unsubscribe = Events.On("workspace-changed", fetchWorkspaces)
    unsubscribeSkills = Events.On("skills-updated", fetchWorkspaces)
  })

  onUnmounted(() => {
    if (unsubscribe) unsubscribe()
    if (unsubscribeSkills) unsubscribeSkills()
  })

  return {
    workspaces,
    activeWorkspace,
    loading,
    error,
    fetchWorkspaces,
    switchWorkspace,
    createWorkspace,
    deleteWorkspace,
  }
}
