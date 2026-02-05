import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import { Dialogs } from "@wailsio/runtime"
import BrowseSkills from "./BrowseSkills.vue"
import { mockAppService } from "../test/setup"
import type { SkillInfo, WorkspaceInfo } from "../types/skill"

describe("BrowseSkills", () => {
  const mockSkills: SkillInfo[] = [
    {
      name: "react-patterns",
      description: "React best practices",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-29T10:00:00Z",
      agents: ["claude-code"],
    },
    {
      name: "typescript-tips",
      description: "TypeScript tips",
      source: "local/path",
      sourceType: "local",
      installedAt: "2025-01-28T10:00:00Z",
      agents: ["claude-code"],
    },
    {
      name: "vue-utils",
      description: "Vue utilities",
      source: "vue/skills",
      sourceType: "github",
      installedAt: "2025-01-27T10:00:00Z",
      agents: ["cursor"],
    },
  ]

  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: "my-workspace",
      description: "My workspace",
      skills: ["react-patterns"],
      isActive: true,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetSkills.mockResolvedValue(mockSkills)
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
  })

  async function mountBrowseSkills() {
    const wrapper = mount(BrowseSkills)
    await flushPromises()
    return wrapper
  }

  describe("rendering", () => {
    it("shows loading state initially", () => {
      mockAppService.GetSkills.mockReturnValue(new Promise(() => {}))
      mockAppService.GetWorkspaces.mockReturnValue(new Promise(() => {}))
      const wrapper = mount(BrowseSkills)

      expect(wrapper.find(".loading").exists()).toBe(true)
    })

    it("shows error on fetch failure", async () => {
      mockAppService.GetSkills.mockRejectedValue(new Error("Failed"))

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".error").exists()).toBe(true)
    })

    it("shows empty state when no skills", async () => {
      mockAppService.GetSkills.mockResolvedValue([])

      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".empty").exists()).toBe(true)
      expect(wrapper.text()).toContain("No skills installed yet")
    })

    it("shows available skills count", async () => {
      const wrapper = await mountBrowseSkills()

      // allSkills is skillsNotInWorkspace, so 2 out of 3
      expect(wrapper.find(".count").text()).toBe("2 skills available")
    })

    it("shows workspace hint", async () => {
      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".workspace-hint").text()).toContain("my-workspace")
    })

    it("shows skills not in workspace with add button", async () => {
      const wrapper = await mountBrowseSkills()

      // Skills not in workspace: typescript-tips, vue-utils
      const mainList = wrapper.find(".skills-list:not(.muted)")
      const cards = mainList.findAllComponents({ name: "SkillCard" })
      expect(cards).toHaveLength(2)
    })

    it("shows skills already in workspace section", async () => {
      const wrapper = await mountBrowseSkills()

      expect(wrapper.find(".in-workspace-section").exists()).toBe(true)
      expect(wrapper.find(".section-label").text()).toContain(
        "Already in workspace (1)",
      )
    })

    it("renders uninstall button on all skills", async () => {
      const wrapper = await mountBrowseSkills()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("showUninstall")).toBe(true)
      })
    })
  })

  describe("add to workspace", () => {
    it("calls AddSkillToWorkspace", async () => {
      mockAppService.AddSkillToWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("add", "typescript-tips")
      await flushPromises()

      expect(mockAppService.AddSkillToWorkspace).toHaveBeenCalledWith(
        "typescript-tips",
        "my-workspace",
      )
    })

    it("shows error on add failure", async () => {
      mockAppService.AddSkillToWorkspace.mockRejectedValue(
        new Error("Add failed"),
      )

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("add", "typescript-tips")
      await flushPromises()

      expect(wrapper.text()).toContain("Add failed")
    })
  })

  describe("uninstall", () => {
    it("calls RemoveSkill on confirm", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Uninstall")
      mockAppService.RemoveSkill.mockResolvedValue(undefined)

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      expect(Dialogs.Question).toHaveBeenCalledWith(
        expect.objectContaining({ Title: "Uninstall Skill" }),
      )
      expect(mockAppService.RemoveSkill).toHaveBeenCalledWith("typescript-tips")
    })

    it("does not uninstall on cancel", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Cancel")

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "typescript-tips")
      await flushPromises()

      expect(mockAppService.RemoveSkill).not.toHaveBeenCalled()
    })

    it("shows error on uninstall failure", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Uninstall")
      mockAppService.RemoveSkill.mockRejectedValue(
        new Error("Uninstall failed"),
      )

      const wrapper = await mountBrowseSkills()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("uninstall", "react-patterns")
      await flushPromises()

      expect(wrapper.text()).toContain("Uninstall failed")
    })
  })
})
