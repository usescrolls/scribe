import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { flushPromises } from "@vue/test-utils"
import { useOnboarding } from "./useOnboarding"
import { mockAppService } from "../test/setup"

describe("useOnboarding", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    mockAppService.IsOnboardingCompleted.mockResolvedValue(false)
    mockAppService.GetAgentStatus.mockResolvedValue([])
    mockAppService.DetectExistingSkills.mockResolvedValue([])
    mockAppService.DetectSkillConflicts.mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe("initial state", () => {
    it("starts on welcome step", () => {
      const { currentStep } = useOnboarding()
      expect(currentStep.value).toBe("welcome")
    })

    it("starts with loading true", () => {
      const { loading } = useOnboarding()
      expect(loading.value).toBe(true)
    })

    it("starts with empty agents", () => {
      const { agents } = useOnboarding()
      expect(agents.value).toEqual([])
    })
  })

  describe("checkOnboardingStatus", () => {
    it("updates isCompleted from API", async () => {
      mockAppService.IsOnboardingCompleted.mockResolvedValue(true)

      const { isCompleted, checkOnboardingStatus } = useOnboarding()
      await checkOnboardingStatus()

      expect(isCompleted.value).toBe(true)
    })

    it("sets error on failure", async () => {
      mockAppService.IsOnboardingCompleted.mockRejectedValue(
        new Error("Failed"),
      )

      const { error, checkOnboardingStatus } = useOnboarding()
      await checkOnboardingStatus()

      expect(error.value).toBe("Failed")
    })

    it("sets loading false after check", async () => {
      const { loading, checkOnboardingStatus } = useOnboarding()
      await checkOnboardingStatus()

      expect(loading.value).toBe(false)
    })
  })

  describe("fetchAgents", () => {
    it("fetches agents from API", async () => {
      const mockAgents = [
        {
          id: "claude-code",
          displayName: "Claude Code",
          installed: true,
          skillCount: 3,
          globalSkillsDir: "",
        },
      ]
      mockAppService.GetAgentStatus.mockResolvedValue(mockAgents)

      const { agents, fetchAgents } = useOnboarding()
      await fetchAgents()

      expect(agents.value).toEqual(mockAgents)
    })

    it("sets agentsLoading during fetch", async () => {
      const { agentsLoading, fetchAgents } = useOnboarding()

      const promise = fetchAgents()
      expect(agentsLoading.value).toBe(true)

      await promise
      expect(agentsLoading.value).toBe(false)
    })
  })

  describe("agent scanning", () => {
    it("starts periodic scanning", async () => {
      const { startAgentScan } = useOnboarding()
      startAgentScan()
      await flushPromises()

      expect(mockAppService.GetAgentStatus).toHaveBeenCalledTimes(1)

      vi.advanceTimersByTime(30000)
      await flushPromises()

      expect(mockAppService.GetAgentStatus).toHaveBeenCalledTimes(2)
    })

    it("stops periodic scanning", async () => {
      const { startAgentScan, stopAgentScan } = useOnboarding()
      startAgentScan()
      await flushPromises()

      stopAgentScan()
      vi.advanceTimersByTime(60000)
      await flushPromises()

      // Only the initial call should have happened
      expect(mockAppService.GetAgentStatus).toHaveBeenCalledTimes(1)
    })
  })

  describe("computed properties", () => {
    it("computes installedAgents", async () => {
      mockAppService.GetAgentStatus.mockResolvedValue([
        {
          id: "a",
          displayName: "A",
          installed: true,
          skillCount: 0,
          globalSkillsDir: "",
        },
        {
          id: "b",
          displayName: "B",
          installed: false,
          skillCount: 0,
          globalSkillsDir: "",
        },
      ])

      const { installedAgents, fetchAgents } = useOnboarding()
      await fetchAgents()

      expect(installedAgents.value).toHaveLength(1)
      expect(installedAgents.value[0].id).toBe("a")
    })

    it("computes hasInstalledAgents", async () => {
      mockAppService.GetAgentStatus.mockResolvedValue([
        {
          id: "a",
          displayName: "A",
          installed: true,
          skillCount: 0,
          globalSkillsDir: "",
        },
      ])

      const { hasInstalledAgents, fetchAgents } = useOnboarding()
      await fetchAgents()

      expect(hasInstalledAgents.value).toBe(true)
    })

    it("computes hasExistingSkills", async () => {
      const { hasExistingSkills, existingSkills } = useOnboarding()
      expect(hasExistingSkills.value).toBe(false)

      existingSkills.value = [
        {
          name: "test",
          path: "/test",
          agentId: "a",
          agentName: "A",
          isGitRepo: false,
        },
      ]
      expect(hasExistingSkills.value).toBe(true)
    })
  })

  describe("fetchExistingSkills", () => {
    it("fetches skills and conflicts", async () => {
      const mockSkills = [
        {
          name: "test",
          path: "/test",
          agentId: "a",
          agentName: "A",
          isGitRepo: false,
        },
      ]
      const mockConflicts = [{ name: "dupe", sources: [] }]
      mockAppService.DetectExistingSkills.mockResolvedValue(mockSkills)
      mockAppService.DetectSkillConflicts.mockResolvedValue(mockConflicts)

      const { existingSkills, skillConflicts, fetchExistingSkills } =
        useOnboarding()
      await fetchExistingSkills()

      expect(existingSkills.value).toEqual(mockSkills)
      expect(skillConflicts.value).toEqual(mockConflicts)
    })

    it("sets loading during fetch", async () => {
      const { existingSkillsLoading, fetchExistingSkills } = useOnboarding()

      const promise = fetchExistingSkills()
      expect(existingSkillsLoading.value).toBe(true)

      await promise
      expect(existingSkillsLoading.value).toBe(false)
    })
  })

  describe("importAllSkills", () => {
    it("calls API and clears skills on success", async () => {
      mockAppService.ImportAllExistingSkills.mockResolvedValue(undefined)

      const { importAllSkills, existingSkills } = useOnboarding()
      existingSkills.value = [
        {
          name: "test",
          path: "/test",
          agentId: "a",
          agentName: "A",
          isGitRepo: false,
        },
      ]

      const result = await importAllSkills()

      expect(result).toBe(true)
      expect(existingSkills.value).toEqual([])
      expect(mockAppService.ImportAllExistingSkills).toHaveBeenCalled()
    })

    it("returns false on failure", async () => {
      mockAppService.ImportAllExistingSkills.mockRejectedValue(
        new Error("Failed"),
      )

      const { importAllSkills, error } = useOnboarding()
      const result = await importAllSkills()

      expect(result).toBe(false)
      expect(error.value).toBe("Failed")
    })
  })

  describe("deleteAllSkills", () => {
    it("calls API and clears skills on success", async () => {
      mockAppService.DeleteAllExistingSkills.mockResolvedValue(undefined)

      const { deleteAllSkills, existingSkills } = useOnboarding()
      existingSkills.value = [
        {
          name: "test",
          path: "/test",
          agentId: "a",
          agentName: "A",
          isGitRepo: false,
        },
      ]

      const result = await deleteAllSkills()

      expect(result).toBe(true)
      expect(existingSkills.value).toEqual([])
    })

    it("returns false on failure", async () => {
      mockAppService.DeleteAllExistingSkills.mockRejectedValue(
        new Error("Failed"),
      )

      const { deleteAllSkills, error } = useOnboarding()
      const result = await deleteAllSkills()

      expect(result).toBe(false)
      expect(error.value).toBe("Failed")
    })
  })

  describe("resolveConflict", () => {
    it("calls API and refreshes skills", async () => {
      mockAppService.ResolveSkillConflict.mockResolvedValue(undefined)

      const { resolveConflict } = useOnboarding()
      const result = await resolveConflict("/path/to/skill")

      expect(result).toBe(true)
      expect(mockAppService.ResolveSkillConflict).toHaveBeenCalledWith(
        "/path/to/skill",
      )
    })

    it("returns false on failure", async () => {
      mockAppService.ResolveSkillConflict.mockRejectedValue(new Error("Failed"))

      const { resolveConflict, error } = useOnboarding()
      const result = await resolveConflict("/path/to/skill")

      expect(result).toBe(false)
      expect(error.value).toBe("Failed")
    })
  })

  describe("installDemoSkill", () => {
    it("calls API on success", async () => {
      mockAppService.InstallDemoSkill.mockResolvedValue(undefined)

      const { installDemoSkill } = useOnboarding()
      const result = await installDemoSkill()

      expect(result).toBe(true)
      expect(mockAppService.InstallDemoSkill).toHaveBeenCalled()
    })

    it("sets installing state during install", async () => {
      const { installDemoSkill, demoSkillInstalling } = useOnboarding()

      const promise = installDemoSkill()
      expect(demoSkillInstalling.value).toBe(true)

      await promise
      expect(demoSkillInstalling.value).toBe(false)
    })

    it("returns false on failure", async () => {
      mockAppService.InstallDemoSkill.mockRejectedValue(new Error("Failed"))

      const { installDemoSkill, error } = useOnboarding()
      const result = await installDemoSkill()

      expect(result).toBe(false)
      expect(error.value).toBe("Failed")
    })
  })

  describe("completeOnboarding", () => {
    it("calls API and sets completed", async () => {
      mockAppService.CompleteOnboarding.mockResolvedValue(undefined)

      const { completeOnboarding, isCompleted } = useOnboarding()
      const result = await completeOnboarding()

      expect(result).toBe(true)
      expect(isCompleted.value).toBe(true)
    })

    it("returns false on failure", async () => {
      mockAppService.CompleteOnboarding.mockRejectedValue(new Error("Failed"))

      const { completeOnboarding, error } = useOnboarding()
      const result = await completeOnboarding()

      expect(result).toBe(false)
      expect(error.value).toBe("Failed")
    })
  })

  describe("navigation", () => {
    it("goToStep changes current step", () => {
      const { goToStep, currentStep } = useOnboarding()

      goToStep("agents")
      expect(currentStep.value).toBe("agents")

      goToStep("complete")
      expect(currentStep.value).toBe("complete")
    })

    it("nextStep advances to next step", () => {
      const { nextStep, currentStep } = useOnboarding()

      nextStep()
      expect(currentStep.value).toBe("agents")

      nextStep()
      expect(currentStep.value).toBe("existing-skills")
    })

    it("nextStep does nothing on last step", () => {
      const { goToStep, nextStep, currentStep } = useOnboarding()

      goToStep("complete")
      nextStep()
      expect(currentStep.value).toBe("complete")
    })

    it("previousStep goes back", () => {
      const { goToStep, previousStep, currentStep } = useOnboarding()

      goToStep("agents")
      previousStep()
      expect(currentStep.value).toBe("welcome")
    })

    it("previousStep does nothing on first step", () => {
      const { previousStep, currentStep } = useOnboarding()

      previousStep()
      expect(currentStep.value).toBe("welcome")
    })
  })
})
