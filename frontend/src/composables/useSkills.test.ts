import { describe, it, expect, vi, beforeEach } from "vitest"
import { flushPromises } from "@vue/test-utils"
import { useSkills } from "./useSkills"
import { mockAppService } from "../test/setup"
import type { SkillInfo } from "../types/skill"

describe("useSkills", () => {
  const mockSkills: SkillInfo[] = [
    {
      name: "react-patterns",
      description: "React best practices",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-29T10:00:00Z",
      agents: ["claude-code", "cursor"],
    },
    {
      name: "typescript-tips",
      description: "TypeScript tips and tricks",
      source: "local/path",
      sourceType: "local",
      installedAt: "2025-01-28T10:00:00Z",
      agents: ["claude-code"],
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetSkills.mockResolvedValue(mockSkills)
  })

  describe("fetchSkills", () => {
    it("fetches skills and updates state", async () => {
      const { skills, loading, fetchSkills } = useSkills()

      await fetchSkills()
      await flushPromises()

      expect(mockAppService.GetSkills).toHaveBeenCalled()
      expect(skills.value).toEqual(mockSkills)
      expect(loading.value).toBe(false)
    })

    it("sets loading state during fetch", async () => {
      const { loading, fetchSkills } = useSkills()

      const fetchPromise = fetchSkills()
      expect(loading.value).toBe(true)

      await fetchPromise
      expect(loading.value).toBe(false)
    })

    it("handles fetch errors", async () => {
      mockAppService.GetSkills.mockRejectedValue(new Error("Network error"))

      const { error, fetchSkills } = useSkills()
      await fetchSkills()

      expect(error.value).toBe("Network error")
    })

    it("clears error on successful fetch", async () => {
      const { error, fetchSkills } = useSkills()

      // First fetch fails
      mockAppService.GetSkills.mockRejectedValueOnce(new Error("Network error"))
      await fetchSkills()
      expect(error.value).toBe("Network error")

      // Second fetch succeeds
      mockAppService.GetSkills.mockResolvedValue(mockSkills)
      await fetchSkills()
      expect(error.value).toBe(null)
    })
  })

  describe("uninstall", () => {
    it("calls RemoveSkill and refreshes list", async () => {
      mockAppService.RemoveSkill.mockResolvedValue(undefined)

      const { uninstall, fetchSkills } = useSkills()
      await fetchSkills()

      const result = await uninstall("react-patterns")

      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("react-patterns")
      expect(result).toBe(true)
      // Should have fetched twice: initial + after uninstall
      expect(mockAppService.GetSkills).toHaveBeenCalledTimes(2)
    })

    it("returns false and sets error on failure", async () => {
      mockAppService.RemoveSkill.mockRejectedValue(
        new Error("Permission denied"),
      )

      const { uninstall, error } = useSkills()
      const result = await uninstall("react-patterns")

      expect(result).toBe(false)
      expect(error.value).toBe("Permission denied")
    })

    it("uses fallback error message for non-Error exceptions", async () => {
      mockAppService.RemoveSkill.mockRejectedValue("Unknown error")

      const { uninstall, error } = useSkills()
      const result = await uninstall("react-patterns")

      expect(result).toBe(false)
      expect(error.value).toBe("Failed to uninstall skill")
    })
  })

  describe("initial state", () => {
    it("starts with empty skills array", () => {
      const { skills } = useSkills()
      expect(skills.value).toEqual([])
    })

    it("starts with loading true", () => {
      const { loading } = useSkills()
      expect(loading.value).toBe(true)
    })

    it("starts with no error", () => {
      const { error } = useSkills()
      expect(error.value).toBe(null)
    })
  })
})
