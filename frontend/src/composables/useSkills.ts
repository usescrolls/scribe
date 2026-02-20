import { ref, onMounted, onUnmounted } from "vue"
import { AppService } from "../bindings/scribe"
import { Events } from "@wailsio/runtime"
import { useLogger } from "./useLogger"
import type { SkillInfo } from "../types/skill"

const log = useLogger("Skills")

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
      log.error(`failed to fetch skills: ${e instanceof Error ? e.message : e}`)
      error.value = e instanceof Error ? e.message : "Failed to load skills"
    } finally {
      loading.value = false
    }
  }

  async function uninstall(name: string): Promise<boolean> {
    log.info(`uninstalling skill: ${name}`)
    try {
      await AppService.RemoveSkill(name)
      log.info(`skill uninstalled: ${name}`)
      await fetchSkills()
      return true
    } catch (e) {
      log.error(
        `failed to uninstall skill ${name}: ${
          e instanceof Error ? e.message : e
        }`,
      )
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
