import { ref, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import type { SkillInfo } from "../types/skill"

export function useSkills() {
  const skills = ref<SkillInfo[]>([])
  const loading = ref(true)
  const error = ref<string | null>(null)
  let unsubscribeSkills: { (): void } | null = null
  let unsubscribeWorkspace: { (): void } | null = null

  async function fetchSkills() {
    try {
      loading.value = true
      error.value = null
      skills.value = await AppService.GetSkills()
    } catch (e) {
      error.value = e instanceof Error ? e.message : "Failed to load skills"
    } finally {
      loading.value = false
    }
  }

  async function uninstall(name: string): Promise<boolean> {
    console.log("[useSkills] uninstall called with:", name)
    try {
      console.log("[useSkills] Calling AppService.RemoveSkill...")
      await AppService.RemoveSkill(name)
      console.log("[useSkills] RemoveSkill succeeded, refreshing skills...")
      await fetchSkills()
      return true
    } catch (e) {
      console.error("[useSkills] RemoveSkill failed:", e)
      error.value = e instanceof Error ? e.message : "Failed to uninstall skill"
      return false
    }
  }

  onMounted(() => {
    fetchSkills()
    unsubscribeSkills = Events.On("skills-updated", fetchSkills)
    unsubscribeWorkspace = Events.On("workspace-changed", fetchSkills)
  })

  onUnmounted(() => {
    if (unsubscribeSkills) {
      unsubscribeSkills()
    }
    if (unsubscribeWorkspace) {
      unsubscribeWorkspace()
    }
  })

  return { skills, loading, error, fetchSkills, uninstall }
}
