import { ref, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import type { WorkspaceInfo } from "../types/skill"

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
      error.value = e instanceof Error ? e.message : "Failed to load workspaces"
    } finally {
      loading.value = false
    }
  }

  async function switchWorkspace(name: string): Promise<boolean> {
    console.log("[useWorkspaces] switching to:", name)
    try {
      await AppService.SetActiveWorkspace(name)
      activeWorkspace.value = name
      await fetchWorkspaces()
      return true
    } catch (e) {
      console.error("[useWorkspaces] switchWorkspace failed:", e)
      error.value =
        e instanceof Error ? e.message : "Failed to switch workspace"
      return false
    }
  }

  async function createWorkspace(
    name: string,
    description: string = "",
  ): Promise<boolean> {
    try {
      await AppService.CreateWorkspace(name, description)
      await fetchWorkspaces()
      return true
    } catch (e) {
      error.value =
        e instanceof Error ? e.message : "Failed to create workspace"
      return false
    }
  }

  async function deleteWorkspace(name: string): Promise<boolean> {
    try {
      await AppService.DeleteWorkspace(name)
      await fetchWorkspaces()
      return true
    } catch (e) {
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
