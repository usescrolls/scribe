import { describe, it, expect, beforeEach } from "vitest"
import { useSkillUpdateChecker } from "./useSkillUpdateChecker"

describe("useSkillUpdateChecker", () => {
  const { sourceUpdates, clearUpdate, hasUpdates, getUpdateInfo } =
    useSkillUpdateChecker()

  beforeEach(() => {
    sourceUpdates.value = {}
  })

  describe("clearUpdate", () => {
    it("clears hasUpdates and updatedSkillNames", () => {
      sourceUpdates.value = {
        "owner/repo": {
          source: "owner/repo",
          hasUpdates: true,
          updatedSkillNames: ["skill-a", "skill-b"],
          newAvailableSkills: [],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      clearUpdate("owner/repo")

      const info = getUpdateInfo("owner/repo")
      expect(info?.hasUpdates).toBe(false)
      expect(info?.updatedSkillNames).toEqual([])
    })

    it("clears newAvailableSkills", () => {
      sourceUpdates.value = {
        "owner/repo": {
          source: "owner/repo",
          hasUpdates: false,
          updatedSkillNames: [],
          newAvailableSkills: [
            {
              name: "new-skill",
              description: "A new skill",
              alreadyInstalled: false,
            },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      clearUpdate("owner/repo")

      const info = getUpdateInfo("owner/repo")
      expect(info?.newAvailableSkills).toEqual([])
    })

    it("does not affect other sources", () => {
      sourceUpdates.value = {
        "owner/repo-a": {
          source: "owner/repo-a",
          hasUpdates: true,
          updatedSkillNames: ["skill-a"],
          newAvailableSkills: [
            { name: "new-a", description: "", alreadyInstalled: false },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
        "owner/repo-b": {
          source: "owner/repo-b",
          hasUpdates: true,
          updatedSkillNames: ["skill-b"],
          newAvailableSkills: [
            { name: "new-b", description: "", alreadyInstalled: false },
          ],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      clearUpdate("owner/repo-a")

      expect(getUpdateInfo("owner/repo-a")?.newAvailableSkills).toEqual([])
      expect(getUpdateInfo("owner/repo-b")?.newAvailableSkills).toHaveLength(1)
      expect(hasUpdates("owner/repo-b")).toBe(true)
    })

    it("is a no-op for unknown source", () => {
      sourceUpdates.value = {
        "owner/repo": {
          source: "owner/repo",
          hasUpdates: true,
          updatedSkillNames: ["skill-a"],
          newAvailableSkills: [],
          checkedAt: "2025-01-30T10:00:00Z",
        },
      }

      clearUpdate("unknown/source")

      expect(hasUpdates("owner/repo")).toBe(true)
    })
  })
})
